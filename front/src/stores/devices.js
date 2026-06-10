import { defineStore } from 'pinia'
import { listDevices, getDeviceData } from '../api/devices'

export const useDeviceStore = defineStore('devices', {
  state: () => ({
    deviceList: [],
    deviceData: {},
    polling: false,
  }),

  actions: {
    async fetchDeviceList() {
      this.deviceList = await listDevices()
    },

    async fetchDeviceData(id) {
      const data = await getDeviceData(id)
      this.deviceData[id] = data
    },

    startPolling(ids, interval = 2000) {
      this.polling = true
      const poll = async () => {
        if (!this.polling) return
        for (const id of ids) {
          await this.fetchDeviceData(id)
        }
        setTimeout(poll, interval)
      }
      poll()
    },

    stopPolling() {
      this.polling = false
    },
  },
})
