/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

import { defineStore } from 'pinia'
import { useConfigStore } from './config'

const equal = function (o1: any, o2: any) {
  return o1.name === o2.name && o1.namespace === o2.namespace
}

export const  useDesktopSessions = defineStore('desktopSession', {

  state: (): {sessions: any[], _audioEnabled:boolean, _recordingEnabled: boolean}  =>  ({
    sessions: [], // deciding against local storage here, but it is still an option
    _audioEnabled: false,
    _recordingEnabled: false
  }),

  actions: {

    setActiveSession (data: any): any  {
      const newSessions: any = []
      this.sessions.forEach((val: any) => {
        if (equal(val, data)) {
          val.active = true
        } else {
          val.active = false
        }
        newSessions.push(val)
      })
      this.sessions = newSessions
    },

    deleteSession ( data: any) {
      this.sessions = this.sessions.filter((val: any) => {
        return !equal(val, data)
      })
      if (this.sessions.length !== 0) {
        this.sessions[0].active = true
      }
    },
    toggleAudio ( data: any) {
      this._audioEnabled = data
    },

    toggleRecording (data: any) {
      this._recordingEnabled = data
    },

    addExistingSession (data: any) {
      data.active = true
      this.sessions.push(data)

    },

    async newSession ( { template, namespace, serviceAccount }: any) {
      try {
        const data: any = { template: template.metadata.name, namespace: namespace }
        if (serviceAccount) {
          data.serviceAccount = serviceAccount
        }

        // TODO
        const session = await useConfigStore().axios.post('/api/sessions', data)
        session.data.template = template

        data.active = true
        this.sessions.push(data)
      
        this.setActiveSession(session.data)

      } catch (err) {
        console.log(`Failed to launch new session from ${template.metadata.name}`)
        console.error(err)
        throw err
      }
    },

    async clearSessions () {
     return Promise.all(this.sessions.map(async (session: any) => {
        await this.deleteSession(session)
      }))
    }

  },

  getters: {
    activeSession: state => state.sessions.filter(sess => sess.active)[0],
    audioEnabled: state => state._audioEnabled,
    recordingEnabled: state => state._recordingEnabled,
    sessionStatus: () => async (data: any ) => {
      try {

        // TODO
        const res = await useConfigStore().axios.get(
          `/api/sessions/${data.namespace}/${data.name}`
        )
        return res.data
      } catch (err) {
        console.log(`Failed to fetch session status for ${data.namespace}/${data.name}`)
        console.error(err)
        throw err
      }
    }
  }

})