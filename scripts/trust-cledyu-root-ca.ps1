[CmdletBinding()]
param(
  [string]$CertificatePath = "",
  [string]$Namespace = "cert-manager",
  [string]$SecretName = "cledyu-root-ca",
  [ValidateSet("CurrentUser", "LocalMachine")]
  [string]$StoreLocation = "CurrentUser",
  [string]$Kubeconfig = ""
)

$ErrorActionPreference = "Stop"

if ($Kubeconfig) {
  $env:KUBECONFIG = $Kubeconfig
}

if (-not $CertificatePath) {
  # 1순위: 레포 안의 PEM (Phase 4 영구화로 git 추적 중. kubectl/KUBECONFIG 의존성 0.)
  $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
  $repoRoot = Resolve-Path (Join-Path $scriptDir "..")
  $repoPem = Join-Path $repoRoot "infra\kubernetes\kubeconfig\cledyu-root-ca.pem"
  if (Test-Path $repoPem) {
    $CertificatePath = $repoPem
  } else {
    # 2순위: cluster 에서 추출 (kubectl 동작 가능한 환경)
    $CertificatePath = Join-Path $env:TEMP "cledyu-root-ca.crt"
    $rootCaBase64 = & kubectl -n $Namespace get secret $SecretName -o "jsonpath={.data.tls\.crt}"
    if ($LASTEXITCODE -ne 0 -or -not $rootCaBase64) {
      throw "Failed to extract $SecretName public certificate from namespace $Namespace."
    }
    [Text.Encoding]::ASCII.GetString([Convert]::FromBase64String($rootCaBase64)) |
      Set-Content -NoNewline -Encoding ascii $CertificatePath
  }
}

$resolvedCertificatePath = (Resolve-Path $CertificatePath).Path
$certificate = [Security.Cryptography.X509Certificates.X509Certificate2]::new(
  $resolvedCertificatePath
)

$certutilArgs = if ($StoreLocation -eq "CurrentUser") {
  @("-user", "-f", "-addstore", "Root", $resolvedCertificatePath)
} else {
  @("-f", "-addstore", "Root", $resolvedCertificatePath)
}

$certutilOutput = & certutil.exe @certutilArgs
if ($LASTEXITCODE -ne 0) {
  $certutilOutput | Write-Host
  throw "Failed to install Cledyu Root CA into $StoreLocation\Root."
}

$store = [Security.Cryptography.X509Certificates.X509Store]::new(
  "Root",
  [Security.Cryptography.X509Certificates.StoreLocation]::$StoreLocation
)
$store.Open([Security.Cryptography.X509Certificates.OpenFlags]::ReadOnly)
try {
  $existing = $store.Certificates.Find(
    [Security.Cryptography.X509Certificates.X509FindType]::FindByThumbprint,
    $certificate.Thumbprint,
    $false
  )
  if ($existing.Count -eq 0) {
    throw "Cledyu Root CA was not found in $StoreLocation\Root after installation."
  }
} finally {
  $store.Close()
}

Write-Host "Cledyu Root CA is trusted in $StoreLocation\Root."
Write-Host ""
Write-Host "Subject    : $($certificate.Subject)"
Write-Host "Issuer     : $($certificate.Issuer)"
Write-Host "Thumbprint : $($certificate.Thumbprint)"
Write-Host "NotAfter   : $($certificate.NotAfter)"
Write-Host ""
Write-Host "Restart Chrome or Edge, then open https://keycloak.cledyu.local/admin/"
