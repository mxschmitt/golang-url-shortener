import toastr from 'toastr'

export default class UtilHelper {
    static deleteEntry(url) {
        fetch(url)
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(() => this.loadRecentURLs())
            .catch(e => e instanceof Promise ? e.then(error => toastr.error(`Could not delete: ${error.error}`)) : null)
    }
};