import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { getSiteTitleMode, setSiteTitleMode, subscribeSiteTitleMode } from "../src/titleModePreference.ts";
import { moveGalleryImage, removeDuplicateGalleryLinks, renameVariantAttribute, toggleVariantImage } from "../src/editorUx.ts";

test("gallery reordering retains links and renormalizes the main image",()=>{
  const original=[
    {url:"https://cdn.example/a.jpg",type:"main"},
    {url:"https://cdn.example/b.jpg",type:"additional"},
    {url:"https://cdn.example/c.jpg",type:"additional"},
  ];
  const moved=moveGalleryImage(original,2,0);
  assert.deepEqual(moved.map(x=>x.url),[original[2].url,original[0].url,original[1].url]);
  assert.deepEqual(moved.map(x=>x.type),["main","additional","additional"]);
  assert.deepEqual(original.map(x=>x.type),["main","additional","additional"]);
});

test("deduplicate only identical gallery URLs and preserve variants/colors",()=>{
  const original=[
    {url:"https://cdn.example/gray.jpg",type:"main"},
    {url:"https://cdn.example/pink.jpg",type:"additional"},
    {url:"https://cdn.example/gray.jpg ",type:"additional"},
    {url:"https://cdn.example/pink2.jpg",type:"additional"},
    {url:"",type:"additional"},
    {url:"",type:"additional"},
  ];
  const result=removeDuplicateGalleryLinks(original);
  assert.deepEqual(result.map(x=>x.url),[
    "https://cdn.example/gray.jpg",
    "https://cdn.example/pink.jpg",
    "https://cdn.example/pink2.jpg",
    "",
    "",
  ]);
  assert.deepEqual(result.map(x=>x.type),["main","additional","additional","additional","additional"]);
  assert.equal(original.length,6);
});

test("variant image checkbox toggles only the selected variant's URLs",()=>{
  const gray=["https://cdn.example/gray.jpg"];
  const pink=["https://cdn.example/pink.jpg"];
  const updated=toggleVariantImage(gray,pink[0],true);
  assert.deepEqual(updated,[gray[0],pink[0]]);
  assert.deepEqual(toggleVariantImage(updated,pink[0],true),updated);
  assert.deepEqual(toggleVariantImage(updated,gray[0],false),pink);
  assert.deepEqual(gray,["https://cdn.example/gray.jpg"]);
});

test("editing a variant attribute name never clobbers another attribute",()=>{
  const original={Color:"Gray",Size:"A"};
  assert.deepEqual(renameVariantAttribute(original,"Color","Colour"),{Colour:"Gray",Size:"A"});
  assert.deepEqual(renameVariantAttribute(original,"Color","Size"),original);
  assert.deepEqual(original,{Color:"Gray",Size:"A"});
});

test("gallery can move an image directly to any position without mutating the source",()=>{
  const original=[
    {url:"https://cdn.example/1.jpg",type:"main"},
    {url:"https://cdn.example/2.jpg",type:"additional"},
    {url:"https://cdn.example/3.jpg",type:"additional"},
    {url:"https://cdn.example/4.jpg",type:"additional"},
  ];
  const moved=moveGalleryImage(original,1,3);
  assert.deepEqual(moved.map(x=>x.url),[original[0].url,original[2].url,original[3].url,original[1].url]);
  assert.deepEqual(moved.map(x=>x.type),["main","additional","additional","additional"]);
  assert.deepEqual(moveGalleryImage(original,2,2),original);
  assert.deepEqual(original.map(x=>x.url),[
    "https://cdn.example/1.jpg","https://cdn.example/2.jpg",
    "https://cdn.example/3.jpg","https://cdn.example/4.jpg",
  ]);
});

test("site title preference defaults safely, persists, survives a failed write, and syncs across windows",t=>{
  const key="una-seller:site-title-mode:v1";
  const values=new Map();
  let failRead=false,failWrite=false;
  const browser=new EventTarget();
  browser.localStorage={
    getItem(k){if(failRead)throw new Error("blocked");return values.get(k)??null;},
    setItem(k,v){if(failWrite)throw new Error("quota");values.set(k,v);},
  };
  const previous=Object.getOwnPropertyDescriptor(globalThis,"window");
  Object.defineProperty(globalThis,"window",{configurable:true,value:browser});
  let unsubscribe=()=>{};
  t.after(()=>{
    unsubscribe();
    if(previous)Object.defineProperty(globalThis,"window",previous);
    else delete globalThis.window;
  });
  assert.equal(getSiteTitleMode(),"localized");
  values.set(key,"invalid");
  assert.equal(getSiteTitleMode(),"localized");
  const observed=[];
  unsubscribe=subscribeSiteTitleMode(()=>observed.push(getSiteTitleMode()));
  setSiteTitleMode("platform_auto");
  assert.equal(values.get(key),"platform_auto");
  assert.deepEqual(observed,["platform_auto"]);
  failWrite=true;
  setSiteTitleMode("localized");
  assert.equal(values.get(key),"platform_auto");
  assert.equal(getSiteTitleMode(),"localized");
  failRead=true;
  assert.equal(getSiteTitleMode(),"localized");
  failRead=false;
  failWrite=false;
  const dispatchStorage=(changedKey,storageArea=browser.localStorage)=>{
    const event=new Event("storage");
    Object.defineProperties(event,{key:{value:changedKey},storageArea:{value:storageArea}});
    browser.dispatchEvent(event);
  };
  dispatchStorage("unrelated");
  assert.deepEqual(observed,["platform_auto","localized"]);
  dispatchStorage(key);
  assert.equal(getSiteTitleMode(),"platform_auto");
  assert.deepEqual(observed,["platform_auto","localized","platform_auto"]);
  values.delete(key);
  dispatchStorage(null);
  assert.equal(getSiteTitleMode(),"localized");
  unsubscribe();
  setSiteTitleMode("platform_auto");
  assert.deepEqual(observed,["platform_auto","localized","platform_auto","localized"]);
  setSiteTitleMode("localized");
});

test("auto mode binds both editors to English without language tabs; localized mode retains all languages",()=>{
  const source=readFileSync(new URL("../src/ListingFieldsEditor.tsx",import.meta.url),"utf8");
  const auto=source.slice(source.indexOf('<fieldset key="platform-auto-title"'),
    source.indexOf('<fieldset key="localized-title-description"'));
  const title=auto.indexOf('{renderTitle("en",');
  const desc=auto.indexOf('<fieldset key="platform-auto-description"');
  assert.ok(title>=0 && desc>title);
  assert.ok(auto.indexOf('{renderDescription("en")}')>desc);
  assert.doesNotMatch(auto,/\{languageTabs\}|render(?:Title|Description)\(activeLanguage/);
  assert.match(source,/key=\{siteTitleMode\+":"\+lang\}/);
  assert.match(source,/key=\{siteTitleMode\+":description:"\+lang\}/);
  assert.match(source,/renderTitle\(activeLanguage,selectedLanguage.full\)/);
  assert.match(source,/renderDescription\(activeLanguage\)/);
  assert.match(source,/englishOnly=\{platformAuto\}/);
  assert.match(source,/if\(platformAuto&&lang!=="en"\)return/);
});

test("AI copy UI hides regional candidates and translation in English-only mode",()=>{
  const source=readFileSync(new URL("../src/AICopyPanel.tsx",import.meta.url),"utf8");
  assert.match(source,/const language = englishOnly \? "en" : selectedLanguage/);
  assert.match(source,/englishOnly && operation === "translate"/);
  assert.match(source,/!englishOnly && <div className="listing-language-tabs"/);
  assert.match(source,/!englishOnly && <button[^>]+!savedEnglish\.trim\(\)/);
  assert.match(source,/!englishOnly && language !== "en" && regional/);
});

test("workbench keeps one editable selling-points area at the bottom and preserves historical descriptions",()=>{
  const source=readFileSync(new URL("../src/ListingFieldsEditor.tsx",import.meta.url),"utf8");
  assert.doesNotMatch(source,/卖点与分类属性|采购成本（人民币 CNY）/);
  assert.match(source,/<legend>附加属性<\/legend>/);
  assert.equal((source.match(/<legend>商品卖点<\/legend>/g)||[]).length,1);
  assert.equal((source.match(/value=\{\(value\.bullet_points\?\?\[\]\)\.join/g)||[]).length,1);
  assert.ok(source.indexOf("<legend>商品卖点</legend>")>source.indexOf("<legend>变体 SKU</legend>"));
  assert.match(source,/set\("bullet_points",points\)/);
  const own=readFileSync(new URL("../src/OwnProductsPage.tsx",import.meta.url),"utf8");
  assert.doesNotMatch(own,/旧版商品描述（历史参考，非默认刊登）/);
  assert.match(own,/description:p\.description/);
  assert.match(own,/updateOwnProduct\(product\.id,\{\.\.\.fields,revision:product\.revision\}/);
  const app=readFileSync(new URL("../src/App.tsx",import.meta.url),"utf8");
  assert.match(app,/product\.own_product_id && product\.bullet_points\.length > 0/);
});
