import { defineStore } from 'pinia'

export const useCanStore = defineStore('can', {
  state: () => ({
    connected: false,
    lastMessage: null,
    messageLog: [],
    socket: null,
    reconnectTimer: null,
  }),

  actions: {
    appendMessage(msg) {
      this.messageLog.push(msg)
      this.lastMessage = msg
    },

    setConnected(status) {
      this.connected = status
    },

    connect() {
      if (this.socket && this.socket.readyState === WebSocket.OPEN) return
      if (this.socket && this.socket.readyState === WebSocket.CONNECTING) return

      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const url = import.meta.env.VITE_CAN_WS_URL || `${protocol}://${window.location.host}/api/v1/ws/can`
      const socket = new WebSocket(url)

      socket.onopen = () => {
        this.setConnected(true)
      }

      socket.onmessage = event => {
        try {
          this.appendMessage(JSON.parse(event.data))
        } catch {
          this.appendMessage({ type: 'raw', data: event.data })
        }
      }

      socket.onclose = () => {
        this.setConnected(false)
        if (this.socket === socket) {
          this.socket = null
        }
        if (!this.reconnectTimer) {
          this.reconnectTimer = setTimeout(() => {
            this.reconnectTimer = null
            this.connect()
          }, 1500)
        }
      }

      socket.onerror = () => {
        this.setConnected(false)
        this.appendMessage({ type: 'ws_error', data: 'WebSocket 连接失败' })
      }

      this.socket = socket
    },

    disconnect() {
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer)
        this.reconnectTimer = null
      }
      if (this.socket) {
        this.socket.close()
        this.socket = null
      }
      this.setConnected(false)
    },
  },
})
