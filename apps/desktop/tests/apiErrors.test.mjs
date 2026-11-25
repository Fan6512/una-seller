import assert from "node:assert/strict";
import test from "node:test";
import { readAPIError } from "../src/apiErrors.ts";

test("missing CBT store shows actionable prerequisite instead of generic 409",async()=>{
  const response=new Response(JSON.stringify({error:"connect a Mercado Libre store",code:"CBT_STORE_REQUIRED"}),{
    status:412,headers:{"content-type":"application/json"},
  });
  const result=await readAPIError(response,"请求失败（HTTP 412）");
  assert.equal(result.code,"CBT_STORE_REQUIRED");
  assert.match(result.message,/设置.*店铺管理/);
});
test("400 field error and converted draft code survive API transport",async()=>{
  const invalid=await readAPIError(new Response(JSON.stringify({error:"purchase_currency must contain 3 uppercase letters"}),{status:400}),"HTTP 400");
  assert.match(invalid.message,/purchase_currency/);
  const converted=await readAPIError(new Response(JSON.stringify({error:"working draft is read-only",code:"WORKING_DRAFT_CONVERTED"}),{status:409}),"HTTP 409");
  assert.equal(converted.code,"WORKING_DRAFT_CONVERTED");
  assert.match(converted.message,/read-only/);
});
test("HTML or oversized remote errors cannot replace bounded fallback",async()=>{
  const malformed=await readAPIError(new Response("<html>proxy error</html>",{status:502}),"代理服务不可用");
  assert.equal(malformed.message,"代理服务不可用");
  const tooLong=await readAPIError(new Response(JSON.stringify({error:"x".repeat(1000)}),{status:500}),"服务器失败");
  assert.equal(tooLong.message,"服务器失败");
});

test("variant upgrade rejection gives actionable guidance through shared API errors",async()=>{
  const result=await readAPIError(new Response(JSON.stringify({code:"VARIANT_CLIENT_UPGRADE_REQUIRED"}),{status:422}),"HTTP 422");
  assert.equal(result.code,"VARIANT_CLIENT_UPGRADE_REQUIRED");
  assert.match(result.message,/复制未保存的修改.*升级.*本次未写入/);
});
