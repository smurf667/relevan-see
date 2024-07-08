#!/usr/bin/env node

import fs from 'fs'
import path from 'path'
import prompt from 'prompt'

async function install(root) {
  let name = undefined;
  switch (process.platform) {
    case 'win32':
      name = 'relevan-see.exe';
      break;
    case 'linux':
      name = 'relevan-see';
      break;
  }
  if (!name) {
    console.error('Platform not supported', process.platform);
    process.exit(1);
  }
  const target = path.join(root, 'hooks', name);
  fs.copyFileSync(path.resolve(process.argv[1], '..', name), target);
  fs.chmodSync(target, 0o755);
  const hook = path.join(root, 'hooks', 'commit-msg');
  let create = true;
  if (fs.existsSync(hook)) {
    const {yesno} = await prompt.get({ description: 'Hook exists. Overwrite (y/n)?', pattern: /^[yn]$/ }, ['yesno']);
    if (yesno === 'n') {
      create = false;
    }
  }
  if (create) {
    fs.writeFileSync(hook, '#!/bin/sh\n./.git/hooks/relevan-see $RELEV_ARGS $1 < /dev/tty\n');
  }
  console.log('relevan-see installed');
}

async function uninstall(root) {
  fs.unlinkSync(path.join(root, 'hooks', 'commit-msg'))
  fs.unlinkSync(path.join(root, 'hooks', `relevan-see${process.platform == "win32" ? ".exe" : "" }`));
  console.log('relevan-see removed');
}

const argv = process.argv.slice(2);

const root = path.join(process.cwd(), '.git');
if (!fs.existsSync(root)) {
  console.error('Must be run from the root of a Git repository, but got:', root);
  process.exit(1);
}

switch (argv.shift()) {
  case 'install':
    install(root);
    break;
  case 'uninstall':
    uninstall(root);
    break;
  default:
    console.error('"install" or "uninstall" argument required');
    break;
}
