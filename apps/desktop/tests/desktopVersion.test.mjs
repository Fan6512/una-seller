import assert from "node:assert/strict";
import test from "node:test";
import { formatDesktopVersion } from "../src/desktopVersion.ts";

test("test candidates show test channel",()=>{
  assert.equal(formatDesktopVersion("0.2.0"),"V0.2 测试版");
  assert.equal(formatDesktopVersion("0.2.1"),"V0.2.1 测试版");
  assert.equal(formatDesktopVersion("0.1.5"),"V0.1.5 测试版");
});
test("stable versions show release channel only when major >= 1",()=>{
  assert.equal(formatDesktopVersion("1.0.0"),"V1.0 正式版");
  assert.equal(formatDesktopVersion("1.6.0"),"V1.6 正式版");
  assert.equal(formatDesktopVersion("2.0.0"),"V2.0 正式版");
});
test("unexpected package version is displayed without guesswork",()=>{
  assert.equal(formatDesktopVersion("0.2.0-beta.1"),"V0.2.0-beta.1");
});
