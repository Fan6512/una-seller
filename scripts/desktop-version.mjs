#!/usr/bin/env node
// Single explicit version change per distributable candidate, never per CI retry.
// Usage: node scripts/desktop-version.mjs --check
//        node scripts/desktop-version.mjs 0.3
//        node scripts/desktop-version.mjs 1.0  # only after release approval
// Package managers and Tauri store 0.3.0/1.0.0, while the product UI shows two parts.
import { readFileSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const root=new URL("../",import.meta.url);
const path=(name)=>fileURLToPath(new URL(name,root));
const sources={
  packageJSON:"apps/desktop/package.json",
  npmLock:"apps/desktop/package-lock.json",
  tauri:"apps/desktop/src-tauri/tauri.conf.json",
  cargo:"apps/desktop/src-tauri/Cargo.toml",
  cargoLock:"apps/desktop/src-tauri/Cargo.lock",
};
const files=Object.fromEntries(Object.entries(sources).map(([k,v])=>[k,readFileSync(path(v),"utf8")]));
const pkg=JSON.parse(files.packageJSON);
const lock=JSON.parse(files.npmLock);
const tauri=JSON.parse(files.tauri);
const cargoVersion=/^version = "([^"]+)"$/m.exec(files.cargo)?.[1];
const cargoLockEntries=[...files.cargoLock.matchAll(/\[\[package\]\]\r?\nname = "una-seller"\r?\nversion = "([^"]+)"/g)];
const cargoLockVersion=cargoLockEntries[0]?.[1];
const versions=[pkg.version,lock.version,lock.packages[""].version,tauri.version,cargoVersion,cargoLockVersion];
const versionFormat=/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;
if(!versions.every(x=>x===versions[0]&&versionFormat.test(x)&&x.endsWith(".0")))
  throw new Error("Desktop version files must agree and use internal MAJOR.MINOR.0: "+JSON.stringify(versions));
const current=versions[0];
if(process.argv[2]==="--check"&&process.argv.length===3){
  console.log("Desktop version files agree:",current);
  process.exit(0);
}
const userVersion=process.argv[2];
const publicFormat=/^(0|[1-9]\d*)\.(0|[1-9]\d*)$/;
if(process.argv.length!==3||!publicFormat.test(userVersion))
  throw new Error("Usage: node scripts/desktop-version.mjs --check | MAJOR.MINOR (e.g. 0.3 or 1.0)");
const target=userVersion+".0";
const parts=(s)=>s.split(".").map(Number);
const [before,after]=[parts(current),parts(target)];
const increasing=after[0]>before[0]||(after[0]===before[0]&&after[1]>before[1]);
if(!increasing)throw new Error("Version must strictly increase: "+current+" -> "+target);
if(before[0]>=1&&after[0]===0)throw new Error("A stable release must not downgrade into 0.x");
const changeJSON=(original,keys)=>{
  const copy=JSON.parse(original);
  for(const key of keys){
    let cursor=copy;
    for(const part of key.slice(0,-1))cursor=cursor[part];
    cursor[key[key.length-1]]=target;
  }
  return JSON.stringify(copy,null,2)+"\n";
};
const cargoName=files.cargo.replace(/^version = "[^"]+"$/m, 'version = "'+target+'"');
if(cargoLockEntries.length!==1)throw new Error("Cargo.lock project package not unique");
const oldLock=cargoLockEntries[0][0];
const cargoLocked=files.cargoLock.replace(oldLock,oldLock.replace(/version = "[^"]+"$/,'version = "'+target+'"'));
const edits={
  packageJSON:changeJSON(files.packageJSON,[["version"]]),
  npmLock:changeJSON(files.npmLock,[["version"],["packages","","version"]]),
  tauri:changeJSON(files.tauri,[["version"]]),
  cargo:cargoName,
  cargoLock:cargoLocked,
};
for(const [name,body] of Object.entries(edits))writeFileSync(path(sources[name]),body);
console.log("Desktop candidate version:",current,"->",target,"(display V"+userVersion+")");
