// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat for the canonical source repository
// @license     https://github.com/chanxuehong/wechat/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"net/http"
	"sync"
)

var _ MessageHandler = (*MessageServeMux)(nil)

// MessageServeMux 实现了一个简单的消息路由器, 同时也是一个 MessageHandler.
type MessageServeMux struct {
	rwmutex               sync.RWMutex
	messageHandlers       map[string]MessageHandler
	eventHandlers         map[string]MessageHandler
	defaultMessageHandler MessageHandler
	defaultEventHandler   MessageHandler
}

func NewMessageServeMux() *MessageServeMux {
	return &MessageServeMux{
		messageHandlers: make(map[string]MessageHandler),
		eventHandlers:   make(map[string]MessageHandler),
	}
}

// 注册 MessageHandler, 处理特定类型的消息.
func (mux *MessageServeMux) MessageHandle(msgType string, handler MessageHandler) {
	if msgType == "" {
		panic("mp: invalid msgType")
	}
	if handler == nil {
		panic("mp: nil handler")
	}

	mux.rwmutex.Lock()
	defer mux.rwmutex.Unlock()

	if mux.messageHandlers == nil {
		mux.messageHandlers = make(map[string]MessageHandler)
	}
	mux.messageHandlers[msgType] = handler
}

// 注册 MessageHandlerFunc, 处理特定类型的消息.
func (mux *MessageServeMux) MessageHandleFunc(msgType string, handler func(http.ResponseWriter, *Request)) {
	mux.MessageHandle(msgType, MessageHandlerFunc(handler))
}

// 注册 MessageHandler, 处理未知类型的消息.
func (mux *MessageServeMux) DefaultMessageHandle(handler MessageHandler) {
	if handler == nil {
		panic("mp: nil handler")
	}

	mux.rwmutex.Lock()
	defer mux.rwmutex.Unlock()

	mux.defaultMessageHandler = handler
}

// 注册 MessageHandlerFunc, 处理未知类型的消息.
func (mux *MessageServeMux) DefaultMessageHandleFunc(handler func(http.ResponseWriter, *Request)) {
	mux.DefaultMessageHandle(MessageHandlerFunc(handler))
}

// 注册 MessageHandler, 处理特定类型的事件.
func (mux *MessageServeMux) EventHandle(eventType string, handler MessageHandler) {
	if eventType == "" {
		panic("mp: invalid eventType")
	}
	if handler == nil {
		panic("mp: nil handler")
	}

	mux.rwmutex.Lock()
	defer mux.rwmutex.Unlock()

	if mux.eventHandlers == nil {
		mux.eventHandlers = make(map[string]MessageHandler)
	}
	mux.eventHandlers[eventType] = handler
}

// 注册 MessageHandlerFunc, 处理特定类型的事件.
func (mux *MessageServeMux) EventHandleFunc(eventType string, handler func(http.ResponseWriter, *Request)) {
	mux.EventHandle(eventType, MessageHandlerFunc(handler))
}

// 注册 MessageHandler, 处理未知类型的事件.
func (mux *MessageServeMux) DefaultEventHandle(handler MessageHandler) {
	if handler == nil {
		panic("mp: nil handler")
	}

	mux.rwmutex.Lock()
	defer mux.rwmutex.Unlock()

	mux.defaultEventHandler = handler
}

// 注册 MessageHandlerFunc, 处理未知类型的事件.
func (mux *MessageServeMux) DefaultEventHandleFunc(handler func(http.ResponseWriter, *Request)) {
	mux.DefaultEventHandle(MessageHandlerFunc(handler))
}

// 获取 msgType 对应的 MessageHandler, 如果没有找到 nil.
func (mux *MessageServeMux) messageHandler(msgType string) (handler MessageHandler) {
	if msgType == "" {
		return nil
	}

	mux.rwmutex.RLock()
	defer mux.rwmutex.RUnlock()

	handler = mux.messageHandlers[msgType]
	if handler != nil {
		return
	}
	return mux.defaultMessageHandler
}

// 获取 eventType 对应的 MessageHandler, 如果没有找到 nil.
func (mux *MessageServeMux) eventHandler(eventType string) (handler MessageHandler) {
	if eventType == "" {
		return nil
	}

	mux.rwmutex.RLock()
	defer mux.rwmutex.RUnlock()

	handler = mux.eventHandlers[eventType]
	if handler != nil {
		return
	}
	return mux.defaultEventHandler
}

// MessageServeMux 实现了 MessageHandler 接口.
func (mux *MessageServeMux) ServeMessage(w http.ResponseWriter, r *Request) {
	if MsgType := r.MixedMsg.MsgType; MsgType == "event" {
		handler := mux.eventHandler(r.MixedMsg.Event)
		if handler == nil {
			return // 返回空串, 符合微信协议
		}
		handler.ServeMessage(w, r)
	} else {
		handler := mux.messageHandler(MsgType)
		if handler == nil {
			return // 返回空串, 符合微信协议
		}
		handler.ServeMessage(w, r)
	}
}
