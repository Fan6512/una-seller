#requires -version 5.1
<#
Issue #57: create synthetic source products in an explicitly isolated local
PostgreSQL QA database, through the real authenticated Go collection API.
Never use with production Go API/DB or a real customer data set.
#>
[CmdletBinding()]
param(
  [switch]$IsolatedDatabaseConfirmed,
  [string]$ApiBase = "http://127.0.0.1:8787",
  [string]$SupabaseUrl = "https://rwktwfxzfkzglxwaulcm.supabase.co",
  [string]$PublishableKey = "sb_publishable_KUDh1isf3-EknWx-rviV0A_w3lm-u0z"
)

$ErrorActionPreference = "Stop"
if (-not $IsolatedDatabaseConfirmed) {
  throw "必须明确指定 -IsolatedDatabaseConfirmed，并先确认本机 Go API 使用隔离测试数据库，而非生产数据库。"
}
if ($ApiBase -ne "http://127.0.0.1:8787" -and $ApiBase -ne "http://localhost:8787") {
  throw "仅允许向本机 8787 端口的 Go API 写入合成 QA 数据。"
}
if (-not $SupabaseUrl.StartsWith("https://")) {
  throw "Supabase Auth URL must use HTTPS."
}
Write-Host "仅检查本机 API：$ApiBase"
$null = Invoke-RestMethod -Method Get -Uri "$ApiBase/healthz" -TimeoutSec 8

$email = Read-Host "请输入单独的测试登录邮箱"
if ([string]::IsNullOrWhiteSpace($email)) { throw "邮箱不能为空。" }
$secure = Read-Host "请输入测试账号密码（不会显示）" -AsSecureString
$bstr = [IntPtr]::Zero
try {
  $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
  $plain = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr)
  $authBody = @{ email = $email.Trim(); password = $plain } | ConvertTo-Json -Compress
} finally {
  if ($bstr -ne [IntPtr]::Zero) {
    [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
  }
  $plain = $null
}
$auth = Invoke-RestMethod -Method Post `
  -Uri "$($SupabaseUrl.TrimEnd('/'))/auth/v1/token?grant_type=password" `
  -ContentType "application/json" -Headers @{ apikey = $PublishableKey } `
  -Body $authBody -TimeoutSec 15
$authBody = $null
if ([string]::IsNullOrWhiteSpace($auth.access_token)) {
  throw "登录服务未返回有效的测试账号 token。"
}
$headers = @{ Authorization = "Bearer $($auth.access_token)" }
# A healthy process may still point to a bad DB; authenticated GET verifies
# the QA process is wired to a usable DB before any synthetic POST.
$null = Invoke-RestMethod -Method Get -Uri "$ApiBase/api/v1/collections?limit=1" `
  -Headers $headers -TimeoutSec 15

$fixtures = @(
  @{
    source_platform = "amazon"; source_site = "US"
    source_url = "https://www.amazon.com/dp/BQA5700001"
    title = "Winter Plant Dry Cleaning Spray for Artificial Wreaths"
    description = "Synthetic QA reference: winter decoration cleaner. Not a real listing."
    price = "12.50"; currency = "USD"; collection_status = "partial"
    bullet_points = @("Synthetic source", "For editor testing only")
  },
  @{
    source_platform = "amazon"; source_site = "US"
    source_url = "https://www.amazon.com/dp/BQA5700002"
    title = "Limpiador para Plantas Artificiales y Guirnaldas"
    description = "Producto de prueba sintético; no es un anuncio real."
    price = "11.25"; currency = "USD"; collection_status = "partial"
    bullet_points = @("Solo para pruebas")
    variants = @(@{ attributes = @{ Color = "Black" }; price = "11.25"; currency = "USD"; image_urls = @() })
  },
  @{
    source_platform = "amazon"; source_site = "US"
    source_url = "https://www.amazon.com/dp/BQA5700003"
    title = "Spray para Limpeza de Plantas Artificiais"
    description = "Produto de teste sintético; não é um anúncio real."
    price = "10.00"; currency = "USD"; collection_status = "partial"
    bullet_points = @("Somente para testes")
  }
)
foreach ($fixture in $fixtures) {
  $body = ConvertTo-Json -InputObject $fixture -Depth 8 -Compress
  $result = Invoke-RestMethod -Method Post -Uri "$ApiBase/api/v1/collections" `
    -ContentType "application/json" -Headers $headers -Body $body -TimeoutSec 20
  if ($result.result -notin @("created", "updated") -or -not $result.collection.id) {
    throw "合成来源未得到确认保存；停止后续写入。"
  }
  Write-Host ("已写入隔离 QA 来源：{0} ({1})" -f $fixture.title, $result.result)
}
$headers = $null
$auth = $null
Write-Host "QA 合成数据已写入本机隔离库，请使用同一测试账号登录 Windows QA 客户端。"
