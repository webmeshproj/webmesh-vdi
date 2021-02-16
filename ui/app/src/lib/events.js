export const Events = Object.freeze({"connected": 1, "disconnected": 2, "update": 3, "error": 4})

export class Emitter {
    constructor() {
        this.callbacks = {}
    }

    on(event, f) {
        if (this.callbacks[event]) {
            this.callbacks[event].push(f)
            return
        }
        this.callbacks[event] = [f]
    }

    emit(event, details) {
        if (this.callbacks[event]) {
            this.callbacks[event].forEach((f) => { f(details) })
        }
    }
}