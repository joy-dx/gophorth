import { create } from 'zustand'

interface Log {
    ref: string
    msg: string
}

export type LogState = {
    logs: Record<string, string[]>
}
export type LogAction = {
    addLog: (log: Log) => void
}
export type LogSlice = LogState & LogAction
export const UseLogStore = create<LogSlice>(
    (set) => ({
        logs: {
            system: [],
        },
        addLog: (log) =>
            set((state) => {
                const { ref, msg } = log
                const existingLogs = state.logs[ref] || []
                return {
                    logs: {
                        ...state.logs,
                        [ref]: [...existingLogs, msg],
                    },
                }
            }),
    })
)