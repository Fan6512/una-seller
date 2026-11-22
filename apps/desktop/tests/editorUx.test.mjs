import test from "node:test";
import assert from "node:assert/strict";
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
