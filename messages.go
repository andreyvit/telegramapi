package telegramapi

import (
	"github.com/andreyvit/telegramapi/mtproto"
)

func (c *Conn) LoadChats(contacts *ContactList) error {
	r, err := c.Send(&mtproto.TLMessagesGetDialogs{
		Flags:      0,
		Limit:      1000,
		OffsetPeer: &mtproto.TLInputPeerEmpty{},
	})
	if err != nil {
		return err
	}
	switch r := r.(type) {
	case *mtproto.TLMessagesDialogs:
		c.updateChats(contacts, r.Dialogs, r.Messages, r.Chats, r.Users)
		return nil
	case *mtproto.TLMessagesDialogsSlice:
		c.updateChats(contacts, r.Dialogs, r.Messages, r.Chats, r.Users)
		return nil
	default:
		return c.HandleUnknownReply(r)
	}
}

func (c *Conn) updateChats(contacts *ContactList, dialogs []*mtproto.TLDialog, messages []mtproto.TLMessageType, chats []mtproto.TLChatType, users []mtproto.TLUserType) {
	c.stateMut.Lock()
	defer c.stateMut.Unlock()

	selfUserID := c.state.UserID

	accessHashByUserID := make(map[int]uint64)
	for _, apiuser := range users {
		if apiuser, ok := apiuser.(*mtproto.TLUser); ok {
			user := contacts.Users[apiuser.ID]
			if user == nil {
				user = &User{ID: apiuser.ID}
				contacts.Users[apiuser.ID] = user
			}
			user.Username = apiuser.Username
			user.FirstName = apiuser.FirstName
			user.LastName = apiuser.LastName
			accessHashByUserID[apiuser.ID] = apiuser.AccessHash
			if user.ID == selfUserID {
				contacts.Self = user
			}
		}
	}

	contacts.Chats = contacts.Chats[:0]

	for _, dialog := range dialogs {
		var chat *Chat
		if upeer, ok := dialog.Peer.(*mtproto.TLPeerUser); ok {
			if user := contacts.Users[upeer.UserID]; user != nil {
				chat = contacts.UserChats[user.ID]
				if chat == nil {
					chat = &Chat{
						Type:     UserChat,
						ID:       user.ID,
						User:     user,
						Messages: newMessageList(),
					}
					contacts.UserChats[user.ID] = chat
				}

				chat.AccessHash = accessHashByUserID[user.ID]
				chat.Username = user.Username

				if user == contacts.Self {
					contacts.SelfChat = chat
				}

				// TODO: dialog.TopMessage
			}
		}
		if chat != nil {
			contacts.Chats = append(contacts.Chats, chat)
		}
	}
}

func (c *Conn) LoadHistory(contacts *ContactList, chat *Chat) error {
	r, err := c.Send(&mtproto.TLMessagesGetHistory{
		Peer:  chat.inputPeer(),
		Limit: 1000,
	})
	if err != nil {
		return err
	}
	switch r := r.(type) {
	case *mtproto.TLMessagesMessages:
		c.updateHistory(contacts, chat, r.Messages, r.Chats, r.Users)
		return nil
	case *mtproto.TLMessagesMessagesSlice:
		c.updateHistory(contacts, chat, r.Messages, r.Chats, r.Users)
		return nil
	default:
		return c.HandleUnknownReply(r)
	}
}

func (c *Conn) updateHistory(contacts *ContactList, chat *Chat, messages []mtproto.TLMessageType, chats []mtproto.TLChatType, users []mtproto.TLUserType) {
}
