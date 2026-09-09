/* eslint-disable @typescript-eslint/no-explicit-any */
import pino, { Logger } from 'pino'

let loggerRoot: Logger

export const loggerInit = () => {
  loggerRoot = pino({
    browser: {
      asObject: true,
      write: {
        info: console.info.bind(console),
        error: console.error.bind(console),
        debug: console.debug.bind(console),
        warn: console.warn.bind(console),
        trace: console.trace.bind(console),
      },
    },
    level: 'debug',
    formatters: {
      level: (label) => {
        return {
          level: label,
        }
      },
    },
  })
}

export const debug = (message: any, objectToShow: any = null) => {
  if (objectToShow !== null) {
    loggerRoot.child(objectToShow).debug(message)
  } else {
    loggerRoot.debug(message)
  }
}

export const error = (message: any, objectToShow: any = null) => {
  if (objectToShow !== null) {
    loggerRoot.child(objectToShow).error(message)
  } else {
    loggerRoot.error(message)
  }
}

export const info = (message: any, objectToShow: any = null) => {
  if (objectToShow !== null) {
    loggerRoot.child(objectToShow).info(message)
  } else {
    loggerRoot.info(message)
  }
}

export const trace = (message: any, objectToShow: any = null) => {
  if (objectToShow !== null) {
    loggerRoot.child(objectToShow).trace(message)
  } else {
    loggerRoot.trace(message)
  }
}

export const warn = (message: any, objectToShow: any = null) => {
  if (objectToShow !== null) {
    loggerRoot.child(objectToShow).warn(message)
  } else {
    loggerRoot.warn(message)
  }
}
