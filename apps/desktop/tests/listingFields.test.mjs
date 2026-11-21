import assert from "node:assert/strict";
import test from "node:test";
import {detectTitleLanguage,initialListingFields,preferredTitle,validateListingFields} from "../src/listingFields.ts";

test("inbox fields begin with captured images and variants, without assigning a fake purchase cost",()=>{
  const source={
    images:[{url:"https://example.com/main.jpg",type:"main"}],
    bullet_points:["Original factual bullet"],
    variants:[{attributes:{颜色:"黑色"},price:"99.90",currency:"USD",image_urls:[]}],
  };
  const working=initialListingFields(null,source);
  assert.equal(working.purchase_price,"");
  assert.equal(working.category_id,"");
  assert.equal(working.images?.[0].url,source.images[0].url);
  assert.equal(working.variants?.[0].attributes.颜色,"黑色");
  working.images[0].url="https://example.com/edited.jpg";
  assert.equal(source.images[0].url,"https://example.com/main.jpg","source remains immutable");
});

test("saved empty image override stays empty and does not silently restore source image",()=>{
  const source={images:[{url:"https://example.com/main.jpg",type:"main"}],bullet_points:[],variants:[]};
  const fields=initialListingFields({images:[],weight_g:"800"},source);
  assert.deepEqual(fields.images,[]);
  assert.equal(fields.weight_g,"800");
});

test("cross-marketplace price and logistics validations preserve missing value semantics",()=>{
  const f=initialListingFields(null,undefined,"Waterproof Winter Plant Cleaning Spray");
  assert.equal(validateListingFields(f),"");
  f.weight_g="1000";
  f.length_cm="20";
  f.site_prices=[{site:"MLB",price:"299.90",currency:"BRL"}];
  assert.equal(validateListingFields(f),"");
  f.weight_g="-1";
  assert.match(validateListingFields(f),/重量/);
  f.weight_g="1000";
  f.site_prices=[{site:"MLB",price:"29,90",currency:"BRL"}];
  assert.match(validateListingFields(f),/售价/);
});


test("capture title is classified into the correct language slot without translating",()=>{
  const en=initialListingFields(null,undefined,"Waterproof Winter Plant Cleaning Spray");
  const es=initialListingFields(null,undefined,"Limpiador para plantas de invierno con aroma");
  const pt=initialListingFields(null,undefined,"Spray para limpeza de plantas com aroma");
  assert.equal(detectTitleLanguage("Waterproof Winter Plant Cleaning Spray"),"en");
  assert.equal(detectTitleLanguage("Limpiador para plantas de invierno con aroma"),"es");
  assert.equal(detectTitleLanguage("Spray para limpeza de plantas com aroma"),"pt");
  assert.equal(en.titles.en,"Waterproof Winter Plant Cleaning Spray");
  assert.equal(en.titles.es,"");
  assert.equal(es.titles.es,"Limpiador para plantas de invierno con aroma");
  assert.equal(es.titles.en,"");
  assert.equal(pt.titles.pt,"Spray para limpeza de plantas com aroma");
  assert.equal(pt.titles.en,"");
  assert.equal(preferredTitle(pt.titles),pt.titles.pt);
});

test("persisted locale titles and edits remain independent and site-free",()=>{
  const fields=initialListingFields({
    titles:{en:"Winter Plant Dry Cleaning Spray",es:"Spray Limpiador de Plantas de Invierno",pt:"Spray Limpa Plantas de Inverno"},
    category_id:"CBT12345",
    site_prices:[],
  },undefined,"Captured English Title");
  assert.equal(fields.titles.en,"Winter Plant Dry Cleaning Spray");
  assert.equal(fields.titles.es,"Spray Limpiador de Plantas de Invierno");
  assert.equal(fields.titles.pt,"Spray Limpa Plantas de Inverno");
  assert.deepEqual(fields.site_prices,[]);
  assert.equal(fields.category_site,"");
  assert.equal(validateListingFields(fields),"");
  fields.category_id="MLB12345";
  assert.match(validateListingFields(fields),/CBT/);
});
