package telegramapi

import (
	"fmt"
	"strings"
	// "sort"
	"time"

	"github.com/andreyvit/telegramapi/mtproto"
)

type ChatType int

const (
	UserChat ChatType = iota
	GroupChat
	ChannelChat
)

var chatTypeStrings = []string{"user", "group", "channel"}

func (t ChatType) String() string {
	return chatTypeStrings[t]
}

type ContactList struct {
	Self *User

	Chats    []*Chat
	SelfChat *Chat

	Users    map[int]*User
	Groups   map[int]*Group
	Channels map[int]*Channel

	UserChats    map[int]*Chat
	GroupChats   map[int]*Chat
	ChannelChats map[int]*Chat
}

func NewContactList() *ContactList {
	return &ContactList{
		Users:    make(map[int]*User),
		Groups:   make(map[int]*Group),
		Channels: make(map[int]*Channel),

		UserChats:    make(map[int]*Chat),
		GroupChats:   make(map[int]*Chat),
		ChannelChats: make(map[int]*Chat),
	}
}

func (contacts *ContactList) FindChatByTitle(title string) *Chat {
	for _, chat := range contacts.Chats {
		if chat.TitleOrName() == title {
			return chat
		}
	}

	var matches int
	var match *Chat
	for _, chat := range contacts.Chats {
		if strings.Contains(chat.TitleOrName(), title) {
			matches++
			match = chat
		}
	}
	if matches == 1 {
		return match
	} else {
		return nil
	}
}

func (contacts *ContactList) FindChatByPeer(peer mtproto.TLPeerType) *Chat {
	switch peer := peer.(type) {
	case *mtproto.TLPeerUser:
		return contacts.UserChats[peer.UserID]
	case *mtproto.TLPeerChat:
		return contacts.GroupChats[peer.ChatID]
	case *mtproto.TLPeerChannel:
		return contacts.ChannelChats[peer.ChannelID]
	default:
		return nil
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

func (chat *Chat) TitleOrName() string {
	if chat.Title != "" {
		return chat.Title
	} else if chat.User != nil {
		return chat.User.Name()
	} else {
		return fmt.Sprintf("Chat %d", chat.ID)
	}
}

func (chat *Chat) inputPeer() mtproto.TLInputPeerType {
	switch chat.Type {
	case UserChat:
		return &mtproto.TLInputPeerUser{UserID: chat.ID, AccessHash: chat.AccessHash}
	case GroupChat:
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

func (user *User) Name() string {
	if user.Username != "" {
		return user.Username
	} else if user.FirstName != "" && user.LastName != "" {
		return user.FirstName + " " + user.LastName
	} else if user.FirstName != "" {
		return user.FirstName
	} else if user.LastName != "" {
		return user.LastName
	} else {
		return fmt.Sprintf("User %d", user.ID)
	}
}

type Channel struct {
	ID    int
	Title string
}

type Group struct {
	ID    int
	Title string

	ParticipantsCount int
}

type MessageList struct {
	TopMessage   *Message
	Messages     []*Message
	MessagesByID map[int]*Message
	MinKnownID   int
}

func newMessageList() *MessageList {
	return &MessageList{
		MessagesByID: make(map[int]*Message),
	}
}

func (messages *MessageList) foundID(id int) {
	if messages.MinKnownID == 0 || messages.MinKnownID > id {
		messages.MinKnownID = id
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

	FwdFrom    *User
	FwdChannel *Chat
	FwdDate    time.Time

	ReplyToID int
	ReplyTo   *Message

	Text string
}

type byMsgDate []*Message

func (a byMsgDate) Len() int           { return len(a) }
func (a byMsgDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byMsgDate) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }
