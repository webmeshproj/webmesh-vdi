export const Events = Object.freeze({"connected": 1, "disconnected": 2, "update": 3, "error": 4})

export class Emitter {
    constructor() {
        this._callbacks = {}
    }

    // on adds an event listener to the extending object
    on(event, f) {
        if (this._callbacks[event]) {
            this._callbacks[event].push(f)
            return
        }
        this._callbacks[event] = [f]
    }

    // emit broadcasts the given event and details on the object
    emit(event, details) {
        if (this._callbacks[event]) {
            this._callbacks[event].forEach((f) => { f(details) })
        }
    }

    // bind will copy the event listeners of the provided object
    // to this one.
    bind(obj) {
        if (!obj._callbacks) { return }
        for ( const [event, cbs] of Object.entries(obj._callbacks)) {
            cbs.forEach((cb) => { this.on(event, cb) })
        }
    }
}