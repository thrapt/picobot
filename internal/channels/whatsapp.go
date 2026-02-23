package channels

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	qrterminal "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/local/picobot/internal/chat"
)

// whatsappLogger adapts the whatsmeow logger to use Go's standard logger
type whatsappLogger struct{}

func (l whatsappLogger) Errorf(msg string, args ...interface{}) {
	log.Printf("[whatsapp] ERROR: "+msg, args...)
}

func (l whatsappLogger) Warnf(msg string, args ...interface{}) {
	log.Printf("[whatsapp] WARN: "+msg, args...)
}

func (l whatsappLogger) Infof(msg string, args ...interface{}) {
	log.Printf("[whatsapp] INFO: "+msg, args...)
}

func (l whatsappLogger) Debugf(msg string, args ...interface{}) {
	// Skip debug logs to reduce noise
}

func (l whatsappLogger) Sub(module string) waLog.Logger {
	return l
}

// quietLogger only logs errors, used during onboarding to keep output clean.
type quietLogger struct{}

func (l quietLogger) Errorf(msg string, args ...interface{}) {
	log.Printf("[whatsapp] ERROR: "+msg, args...)
}
func (l quietLogger) Warnf(msg string, args ...interface{})  {}
func (l quietLogger) Infof(msg string, args ...interface{})  {}
func (l quietLogger) Debugf(msg string, args ...interface{}) {}
func (l quietLogger) Sub(module string) waLog.Logger         { return l }

// StartWhatsApp starts a WhatsApp bot using the whatsmeow library.
// dbPath is the path to the SQLite database for storing session data.
// allowFrom restricts which phone numbers (in international format, e.g., "1234567890") may send messages; empty means allow all.
func StartWhatsApp(ctx context.Context, hub *chat.Hub, dbPath string, allowFrom []string) error {
	if dbPath == "" {
		return fmt.Errorf("whatsapp database path not provided")
	}

	// Ensure the database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return fmt.Errorf("failed to create whatsapp db directory: %w", err)
	}

	// Initialize database container
	container, err := sqlstore.New(ctx, "sqlite3", "file:"+dbPath+"?_foreign_keys=on", whatsappLogger{})
	if err != nil {
		return fmt.Errorf("failed to connect to whatsapp database: %w", err)
	}

	// Get the first device from the store, or create a new one
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get whatsapp device: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, whatsappLogger{})

	// Check if already logged in
	if client.Store.ID == nil {
		return fmt.Errorf("whatsapp not authenticated - please run 'picobot onboard whatsapp' first")
	}

	// Build allowlist
	allowed := make(map[string]struct{}, len(allowFrom))
	for _, num := range allowFrom {
		allowed[num] = struct{}{}
	}

	waClient := &whatsappClient{
		client:     client,
		hub:        hub,
		outCh:      hub.Subscribe("whatsapp"),
		allowed:    allowed,
		ctx:        ctx,
		typingStop: make(map[string]chan struct{}),
	}

	// Register event handler
	client.AddEventHandler(waClient.handleEvent)

	// Connect to WhatsApp
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to whatsapp: %w", err)
	}

	log.Printf("whatsapp: connected as %s", client.Store.ID.User)

	// Start outbound message handler
	go waClient.runOutbound()

	// Handle disconnection on context cancellation
	go func() {
		<-ctx.Done()
		log.Println("whatsapp: shutting down")
		waClient.stopAllTyping()
		client.Disconnect()
	}()

	return nil
}

// SetupWhatsApp displays a QR code for WhatsApp authentication.
// This should be run once to authenticate the device before starting the bot.
func SetupWhatsApp(dbPath string) error {
	if dbPath == "" {
		return fmt.Errorf("whatsapp database path not provided")
	}

	ctx := context.Background()

	// Ensure the database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return fmt.Errorf("failed to create whatsapp db directory: %w", err)
	}

	// Use quiet logger to keep onboarding output clean
	container, err := sqlstore.New(ctx, "sqlite3", "file:"+dbPath+"?_foreign_keys=on", quietLogger{})
	if err != nil {
		return fmt.Errorf("failed to connect to whatsapp database: %w", err)
	}

	// Get the first device from the store, or create a new one
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get whatsapp device: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, quietLogger{})

	// Check if already logged in
	if client.Store.ID != nil {
		fmt.Printf("Already authenticated as %s\n", client.Store.ID.User)
		fmt.Println("To re-authenticate, delete the database file and run setup again.")
		return nil
	}

	// Listen for the Connected event that fires after the post-pairing
	// 515 reconnect completes.
	connected := make(chan struct{}, 1)
	client.AddEventHandler(func(evt interface{}) {
		if _, ok := evt.(*events.Connected); ok {
			select {
			case connected <- struct{}{}:
			default:
			}
		}
	})

	qrChan, _ := client.GetQRChannel(context.Background())

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to whatsapp: %w", err)
	}
	defer client.Disconnect()

	fmt.Println("Scan the QR code below with WhatsApp on your phone:")
	fmt.Println("(Open WhatsApp > Settings > Linked Devices > Link a Device)")
	fmt.Println()

	for evt := range qrChan {
		if evt.Event == "code" {
			// Display QR code in terminal
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			fmt.Println()
		} else if evt.Event == "success" {
			fmt.Println("Pairing successful, finishing setup...")
		} else if evt.Event == "timeout" {
			return fmt.Errorf("QR code timed out, please try again")
		}
	}

	// Wait for the post-pairing reconnection, then hold the connection
	// so WhatsApp can complete the initial device sync.
	select {
	case <-connected:
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timed out waiting for connection after pairing")
	}

	// Give WhatsApp time to finish the initial sync with the phone.
	fmt.Println("Syncing with phone...")
	time.Sleep(15 * time.Second)

	fmt.Println("Successfully authenticated!")
	if client.Store.ID != nil {
		fmt.Printf("Logged in as: %s\n", client.Store.ID.User)
	}
	return nil
}

// whatsappClient handles WhatsApp messaging
type whatsappClient struct {
	client     *whatsmeow.Client
	hub        *chat.Hub
	outCh      <-chan chat.Outbound
	allowed    map[string]struct{}
	ctx        context.Context
	typingMu   sync.Mutex
	typingStop map[string]chan struct{}
}

// handleEvent processes WhatsApp events
func (c *whatsappClient) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected, *events.PushNameSetting:
		// Mark ourselves as online once the connection is fully established.
		// This is required for chat presence (composing indicator) to work.
		if err := c.client.SendPresence(c.ctx, types.PresenceAvailable); err != nil {
			log.Printf("whatsapp: failed to send available presence: %v", err)
		}
		_ = v
	case *events.Message:
		c.handleMessage(v)
	}
}

// handleMessage processes incoming WhatsApp messages
func (c *whatsappClient) handleMessage(msg *events.Message) {
	// Skip messages from ourselves
	if msg.Info.IsFromMe {
		return
	}

	// Skip group messages (only handle direct messages)
	if msg.Info.IsGroup {
		return
	}

	// Extract sender phone number (without @ domain)
	senderID := msg.Info.Sender.User

	// Enforce allowlist when one is configured
	if len(c.allowed) > 0 {
		if _, ok := c.allowed[senderID]; !ok {
			log.Printf("whatsapp: dropped message from unauthorized user %s", senderID)
			return
		}
	}

	// Mark the message as read (blue ticks) so the sender's phone gets a notification.
	_ = c.client.MarkRead(c.ctx, []types.MessageID{msg.Info.ID}, msg.Info.Timestamp, msg.Info.Chat, msg.Info.Sender)

	// Extract message content
	content := ""
	if msg.Message.Conversation != nil {
		content = *msg.Message.Conversation
	} else if msg.Message.ExtendedTextMessage != nil && msg.Message.ExtendedTextMessage.Text != nil {
		content = *msg.Message.ExtendedTextMessage.Text
	}

	// Handle image messages
	if msg.Message.ImageMessage != nil {
		if msg.Message.ImageMessage.Caption != nil {
			content = *msg.Message.ImageMessage.Caption
		}
		content += "\n[Image received - images not yet supported]"
	}

	// Handle document messages
	if msg.Message.DocumentMessage != nil {
		if msg.Message.DocumentMessage.Caption != nil {
			content = *msg.Message.DocumentMessage.Caption
		}
		if msg.Message.DocumentMessage.FileName != nil {
			content += fmt.Sprintf("\n[Document: %s - documents not yet supported]", *msg.Message.DocumentMessage.FileName)
		}
	}

	if content == "" {
		return
	}

	content = strings.TrimSpace(content)
	chatID := msg.Info.Chat.String()

	log.Printf("whatsapp: received message from %s in chat %s: %s", senderID, chatID, truncate(content, 50))

	c.startTyping(msg.Info.Chat)

	c.hub.In <- chat.Inbound{
		Channel:   "whatsapp",
		SenderID:  senderID,
		ChatID:    chatID,
		Content:   content,
		Timestamp: msg.Info.Timestamp,
		Metadata: map[string]interface{}{
			"message_id": msg.Info.ID,
			"is_group":   msg.Info.IsGroup,
		},
	}
}

// runOutbound reads replies from the hub's whatsapp subscription and sends them
func (c *whatsappClient) runOutbound() {
	for {
		select {
		case <-c.ctx.Done():
			log.Println("whatsapp: stopping outbound sender")
			return
		case out := <-c.outCh:
			// Parse the chat ID (should be in JID format)
			recipient, err := types.ParseJID(out.ChatID)
			if err != nil {
				log.Printf("whatsapp: invalid chat ID %s: %v", out.ChatID, err)
				continue
			}

			c.stopTyping(out.ChatID)

			// Split long messages if needed (WhatsApp limit is around 65KB but we'll use 4096 for safety)
			chunks := splitMessage(out.Content, 4096)

			for i, chunk := range chunks {
				msg := &waProto.Message{
					Conversation: &chunk,
				}

				_, err := c.client.SendMessage(c.ctx, recipient, msg)
				if err != nil {
					log.Printf("whatsapp: send error (chunk %d): %v", i+1, err)
				}
			}
		}
	}
}

// startTyping begins (or resets) a continuous "composing" presence for a chat.
// It stops automatically after 5 minutes or when stopTyping / stopAllTyping is called.
func (c *whatsappClient) startTyping(jid types.JID) {
	key := jid.String()
	c.typingMu.Lock()
	if stop, ok := c.typingStop[key]; ok {
		close(stop)
	}
	stop := make(chan struct{})
	c.typingStop[key] = stop
	c.typingMu.Unlock()

	go func() {
		_ = c.client.SendChatPresence(c.ctx, jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)

		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()
		timeout := time.NewTimer(5 * time.Minute)
		defer timeout.Stop()

		for {
			select {
			case <-stop:
				_ = c.client.SendChatPresence(c.ctx, jid, types.ChatPresencePaused, types.ChatPresenceMediaText)
				return
			case <-timeout.C:
				return
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				_ = c.client.SendChatPresence(c.ctx, jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)
			}
		}
	}()
}

// stopTyping cancels the typing indicator for the given chat.
func (c *whatsappClient) stopTyping(chatID string) {
	c.typingMu.Lock()
	defer c.typingMu.Unlock()
	if stop, ok := c.typingStop[chatID]; ok {
		close(stop)
		delete(c.typingStop, chatID)
	}
}

// stopAllTyping cancels all active typing indicators.
func (c *whatsappClient) stopAllTyping() {
	c.typingMu.Lock()
	defer c.typingMu.Unlock()
	for _, stop := range c.typingStop {
		close(stop)
	}
	c.typingStop = make(map[string]chan struct{})
}
