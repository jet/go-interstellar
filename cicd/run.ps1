<#
SPDX-License-Identifier: Apache-2.0

Copyright (c) 2019-present, Jet.com, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http:#www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License."
#>

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# Set CosmosDB Data + Sources Directories
if (-not (Test-Path env:COSMOSDB_DATA_PATH)) { $env:COSMOSDB_DATA_PATH="${env:LOCALAPPDATA}\CosmosDBEmulator\bind-mount" }
if (-not (Test-Path env:SOURCES)) { $env:SOURCES=[System.IO.Path]::GetFullPath( (Join-Path "$PSScriptRoot" "..\") ) }

# Ensure cosmosdb data path is clean
Remove-Item -Path "$env:COSMOSDB_DATA_PATH" -Recurse -ErrorAction Ignore | Out-Null
New-Item -ItemType directory -Path "$env:COSMOSDB_DATA_PATH" -ErrorAction Stop | Out-Null

# Install-Go
if(-not (Test-Path -Path "env:GOROOT")) {
    Write-Host "=== Installing GO ==="
    $downloadUrl = "https://dl.google.com/go/go1.12.5.windows-amd64.zip"
    $sha256hash  = "ccb694279aab39fe0e70629261f13b0307ee40d2d5e1138ed94738023ab04baa"

    Write-Host "Downloading Go"
    Invoke-WebRequest -Uri $downloadURL -OutFile go.zip

    If((Get-FileHash go.zip -Algorithm sha256).Hash -ne $sha256hash) {
        Write-Host "Failed to validate go.zip"
        exit 1
    }
    
    $env:GOROOT = "$env:BUILD_BINARIESDIRECTORY\go"
    $env:PATH = "${env:PATH};${env:GOROOT}\bin"
    
    Write-Host "Expanding go.zip..."
    Expand-Archive go.zip -DestinationPath $env:BUILD_BINARIESDIRECTORY
    
    Write-Host "Validating Go Version"
    go version

    Remove-Item go.zip
    Write-Host "Compelete"
}


Write-Host "=== Initialized ==="
Write-Host " - Cosmos DB Data : '$env:COSMOSDB_DATA_PATH'"
Write-Host " - Sources        : '$env:SOURCES'"

Write-Host "=== Starting CosmosDB ==="
& docker-compose --compatibility up --detach

Write-Host "=== Running Test ==="
& powershell.exe ./test.ps1
$result = $LASTEXITCODE

Write-Host "=== Cleaning Up ==="
& docker-compose --compatibility down

Write-Host "### Exiting with: $result"
exit $result