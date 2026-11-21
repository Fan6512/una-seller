import assert from "node:assert/strict";
import test from "node:test";
import {initialListingFields,validateListingFields} from "../src/listingFields.ts";

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
  const f=initialListingFields();
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
