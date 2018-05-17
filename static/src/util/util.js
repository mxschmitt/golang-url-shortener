import toastr from 'toastr'

export default class UtilHelper {
    static deleteEntry(url, cb) {
        fetch(url, {credentials: "include"})
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(cb())
            .catch(e => this._reportError(e, "delete entry"))
    }
    static _constructFetch(url, body, cbSucc, cbErr) {
        fetch(url, {
            method: "POST",
            credentials: "include",
            body: JSON.stringify(body),
            headers: {
                'Authorization': window.localStorage.getItem('token'),
                'Content-Type': 'application/json'
            }
        })
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(res => cbSucc ? cbSucc(res) : null)
            .catch(e => {
                if (cbErr) {
                    cbErr(e)
                } else {
                    let name = url.split("/").pop()
                    this._reportError(e, name)
                }
            })
    }
    static _reportError(e, name) {
        if (e instanceof Promise) {
            e.then(error => toastr.error(`Could not fetch ${name}: ${error.error}`))
        } else {
            toastr.error(`Could not fetch ${name}: ${e}`)
        }
    }
    static lookupEntry(ID, cbSucc, cbErr) {
        this._constructFetch("/api/v1/protected/lookup", { ID }, cbSucc, cbErr)
    }
    static getVisitors(ID, cbSucc) {
        this._constructFetch("/api/v1/protected/visitors", { ID }, cbSucc)
    }
    static createEntry(entry, cbSucc) {
        this._constructFetch("/api/v1/protected/create",entry, cbSucc)
    }
    static getRecentURLs(cbSucc) {
        fetch('/api/v1/protected/recent', {
            credentials: "include",
            headers: {
                'Authorization': window.localStorage.getItem('token'),
                'Content-Type': 'application/json'
            }
        })
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(res => cbSucc ? cbSucc(res) : null)
            .catch(e => this._reportError(e, "recent"))
    }
}
