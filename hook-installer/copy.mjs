#!/usr/bin/env node

import fs from 'fs'

const base = '../tool/bin';
if (fs.existsSync(base)) {
  for (const name of fs.readdirSync(base)) {
    fs.copyFileSync(`${base}/${name}`, `bin/${name}`);
  }
}