# relevan-see hook installer

A node-based installer for the relevan-see hook.

The tool can be run like this in the root of a Git repository

	npx relevan-see install

The hook can be removed by calling

	npx relevan-see uninstall

## Building

To build the installer run

```sh
npm install
npm run copy
npm pack
```

The [tool](../tool/) has to have been built before.
