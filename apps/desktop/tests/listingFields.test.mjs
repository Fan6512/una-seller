import assert from "node:assert/strict";
import test from "node:test";
import {detectTitleLanguage,initialListingFields,listingTextCount,preferredTitle,validateListingFields} from "../src/listingFields.ts";

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

test("shared purchase cost and logistics remain independent of target marketplaces",()=>{
  const f=initialListingFields(null,undefined,"Waterproof Winter Plant Cleaning Spray");
  assert.equal(validateListingFields(f),"");
  f.weight_g="1000";
  f.length_cm="20";
  f.purchase_price="29.90";
  f.purchase_currency="CNY";
  assert.equal(validateListingFields(f),"");
  assert.equal("site_prices" in f,false);
  assert.equal("category_site" in f,false);
  f.weight_g="-1";
  assert.match(validateListingFields(f),/重量/);
  f.weight_g="1000";
  f.purchase_price="29,90";
  assert.match(validateListingFields(f),/采购成本/);
});


test("capture title is classified into the correct language slot without translating",()=>{
  const en=initialListingFields(null,undefined,"Waterproof Winter Plant Cleaning Spray");
  const es=initialListingFields(null,undefined,"Limpiador para plantas de invierno con aroma");
  const pt=initialListingFields(null,undefined,"Spray para limpeza de plantas com aroma");
  assert.equal(detectTitleLanguage("Waterproof Winter Plant Cleaning Spray"),"en");
  assert.equal(detectTitleLanguage("Limpiador para plantas de invierno con aroma"),"es");
  assert.equal(detectTitleLanguage("Spray para limpeza de plantas com aroma"),"pt");
  assert.equal(detectTitleLanguage("Limpiador para plantas"),"es");
  assert.equal(detectTitleLanguage("Spray para limpeza de plantas"),"pt");
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
  },undefined,"Captured English Title");
  assert.equal(fields.titles.en,"Winter Plant Dry Cleaning Spray");
  assert.equal(fields.titles.es,"Spray Limpiador de Plantas de Invierno");
  assert.equal(fields.titles.pt,"Spray Limpa Plantas de Inverno");
  assert.equal("site_prices" in fields,false);
  assert.equal("category_site" in fields,false);
  assert.equal(validateListingFields(fields),"");
  fields.category_id="MLB12345";
  assert.match(validateListingFields(fields),/CBT/);
});

test("currency and attribute validation matches server contract before saving",()=>{
  const f=initialListingFields(null,undefined,"Waterproof Winter Plant Cleaning Spray");
  f.purchase_price="18.50";
  f.purchase_currency="cny";
  assert.match(validateListingFields(f),/币种/);
  f.purchase_currency="CNY";
  assert.equal(validateListingFields(f),"");
  f.variants=[{seller_sku:"SKU-1",attributes:{颜色:"黑色"},currency:"usd",price:"19.90",image_urls:[]}];
  assert.match(validateListingFields(f),/币种/);
  f.variants[0].currency="USD";
  assert.equal(validateListingFields(f),"");
  f.attributes=[{name:"a".repeat(121),value:"yes"}];
  assert.match(validateListingFields(f),/属性/);
  f.attributes=[{name:"model",value:"v".repeat(501)}];
  assert.match(validateListingFields(f),/属性/);
  f.attributes=[];
  f.variants[0].image_urls=["http://not-https.example/bad.jpg"];
  assert.match(validateListingFields(f),/URL/);
});


test("a variant needs at least one complete attribute before saving",()=>{
  const f=initialListingFields(null,undefined,"Waterproof Winter Plant Cleaning Spray");
  f.variants=[{seller_sku:"",attributes:{},price:"",currency:"",image_urls:[]}];
  assert.match(validateListingFields(f),/变体.*属性/);
  f.variants[0].attributes={"属性1":""};
  assert.match(validateListingFields(f),/变体.*属性/);
  f.variants[0].attributes={颜色:"黑色"};
  assert.equal(validateListingFields(f),"");
});

test("purchase drafts default to CNY without relabeling historical non-CNY costs",()=>{
  const fresh=initialListingFields(null,undefined,"Example product title");
  assert.equal(fresh.purchase_currency,"CNY");
  assert.equal(fresh.purchase_price,"");
  const oldUSD=initialListingFields({purchase_price:"12.00",purchase_currency:"USD"},undefined,"Example product title");
  assert.equal(oldUSD.purchase_currency,"USD");
  assert.equal(oldUSD.purchase_price,"12.00");
  const missingCurrency=initialListingFields({purchase_price:"12.00",purchase_currency:""},undefined,"Example product title");
  assert.equal(missingCurrency.purchase_currency,"");
  assert.equal(missingCurrency.purchase_price,"12.00");
});

test("regional titles count NFC Unicode characters and preserve unchanged saved titles", () => {
  const fields = initialListingFields(null, undefined, "Glass jar");
  fields.titles.es = "e\u0301".repeat(59) + "!";
  fields.titles.pt = "🫗".repeat(59) + " ";
  assert.equal(listingTextCount(fields.titles.es), 60);
  assert.equal(listingTextCount(fields.titles.pt), 60);
  assert.equal(validateListingFields(fields), "");
  fields.titles.pt += ".";
  const previous = { ...fields.titles };
  assert.match(validateListingFields(fields), /60/);
  assert.equal(validateListingFields(fields, previous), "");
  fields.titles.pt = fields.titles.pt.replace(".", ",");
  assert.match(validateListingFields(fields, previous), /60/);
});

test("optional localized descriptions round-trip independently as multiline plain text", () => {
  const legacy = initialListingFields(null, undefined, "Glass jar");
  assert.equal(Object.hasOwn(legacy, "descriptions"), false);
  const stored = { en: "A glass jar.\n\nFor indoor use.", es: "Para guardar.", pt: "🧪".repeat(20000) };
  const fields = initialListingFields({ ...legacy, descriptions: stored });
  assert.deepEqual(fields.descriptions, stored);
  assert.equal(validateListingFields(fields), "");
  for (const invalid of ["<p>Hola</p>", "**Hola**", "# Hola", "- Hola", "[Hola](https://example.com)", "\x00"]) {
    fields.descriptions.es = invalid;
    assert.match(validateListingFields(fields), /纯文本/);
  }
  fields.descriptions.es = "Para guardar.";
  fields.descriptions.pt += "🧪";
  assert.match(validateListingFields(fields), /20000/);
  assert.equal(stored.pt.length, 40000);
});
