import {EventsOn} from '@wails/runtime'
import {debug, error, info, warn, UseLogStore} from '@lib'
import {main} from '@wailsmodels'

const { Channel, Relay } = main

export interface RuntimeEvent {
  ref: string
  level: string
  timestamp: string
  data: any // eslint-disable-line @typescript-eslint/no-explicit-any
}


const pushLogEvent = (msg: RuntimeEvent) => {
  const logPush = UseLogStore.getState().addLog
  if ('data' in msg && 'msg' in msg.data) {
    logPush({ ref: 'system', msg: msg.data['msg'] })
    switch (msg.level) {
      case 'debug':
        debug(msg.data['msg'])
        break
      case 'error':
        error(msg.data['msg'])
        break
      case 'info':
        info(msg.data['msg'])
        break
      case 'warn':
        warn(msg.data['msg'])
        break
      default:
        logPush({ ref: 'system', msg: `wails_listener_log: unhandled message level ${msg.level}` })
    }
  } else {
      console.log(msg)
  }
}

const handleBaseEvent = (msg: RuntimeEvent) => {
  pushLogEvent(msg)
}

const handleNetEvent = (msg: RuntimeEvent) => {
  pushLogEvent(msg)
}

const handleReleaserEvent = (msg: RuntimeEvent) => {
  pushLogEvent(msg)
}

const handleUpdaterEvent = (msg: RuntimeEvent) => {
  pushLogEvent(msg)
}

EventsOn(Channel.RELAY_BASE, handleBaseEvent)
EventsOn(Channel.RELAY_NET, handleNetEvent)
EventsOn(Channel.RELAY_RELEASER, handleReleaserEvent)
EventsOn(Channel.RELAY_UPDATER, handleUpdaterEvent)