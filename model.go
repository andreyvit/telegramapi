package telegramapi

import (
	"github.com/andreyvit/telegramapi/mtproto"
	"time"
)

type ChatType int

const (
	UserChat ChatType = iota
	ChatChat
	ChannelChat
)

type ContactList struct {
	Self *User

	Chats    []*Chat
	SelfChat *Chat

	Users        map[int]*User
	UserChats    map[int]*Chat
	ChatChats    map[int]*Chat
	ChannelChats map[int]*Chat
}

func NewContactList() *ContactList {
	return &ContactList{
		Users: make(map[int]*User),

		UserChats:    make(map[int]*Chat),
		ChatChats:    make(map[int]*Chat),
		ChannelChats: make(map[int]*Chat),
	}
}

type Chat struct {
	Type ChatType

	// user ID, chat ID or channnel ID
	ID int

	User *User

	// valid for users and channels
	AccessHash uint64

	// valid for chats and channels
	Title string

	// Date time.Time

	Messages *MessageList

	// valid for channels and users
	Username string
}

func (chat *Chat) inputPeer() mtproto.TLInputPeerType {
	switch chat.Type {
	case UserChat:
		return &mtproto.TLInputPeerUser{UserID: chat.ID, AccessHash: chat.AccessHash}
	case ChatChat:
		return &mtproto.TLInputPeerChat{ChatID: chat.ID}
	case ChannelChat:
		return &mtproto.TLInputPeerChannel{ChannelID: chat.ID, AccessHash: chat.AccessHash}
	default:
		panic("unexpected chat type")
	}
}

type User struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
}

type MessageList struct {
	TopMessage   *Message
	Messages     []*Message
	MessagesByID map[int]*Message
}

func newMessageList() *MessageList {
	return &MessageList{
		MessagesByID: make(map[int]*Message),
	}
}

type MessageType int

const (
	NormalMessage MessageType = iota
)

type Message struct {
	ID   int
	Type MessageType

	Date     time.Time
	EditDate time.Time

	From *User
	To   *User

	FwdFrom    *User
	FwdChannel *Chat
	FwdDate    time.Time

	Text string
}
