# kvdi (app)

A Kubernetes-Native VDI

## Install the dependencies
```bash
yarn
```

### Start the app in development mode (hot-code reloading, error reporting, etc.)

Make sure you have the quasar CLI installed.

The `quasar.conf.js` assumes a local instance of the API accessible on `https://localhost:8443`.

If you set up a test environment using the `Makefile` at the root of the repository,
then you should be able to run `make forward-app` to create the required tunnel.

```bash
quasar dev
```

### Lint the files
```bash
yarn run lint
```

### Build the app for production
```bash
quasar build
```

### Customize the configuration
See [Configuring quasar.conf.js](https://quasar.dev/quasar-cli/quasar-conf-js).
