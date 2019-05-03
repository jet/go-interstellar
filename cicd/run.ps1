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

# Set CosmosDB Data + Sources Directories
if (-not (Test-Path env:COSMOSDB_DATA_PATH)) { $env:COSMOSDB_DATA_PATH="${env:LOCALAPPDATA}\CosmosDBEmulator\bind-mount" }
if (-not (Test-Path env:SOURCES)) { $env:SOURCES=[System.IO.Path]::GetFullPath( (Join-Path "$PSScriptRoot" "..\") ) }

# Ensure cosmosdb data path is clean
Remove-Item -Path "$env:COSMOSDB_DATA_PATH" -Recurse -ErrorAction Ignore | Out-Null
New-Item -ItemType directory -Path "$env:COSMOSDB_DATA_PATH" -ErrorAction Stop | Out-Null

Write-Host "=== Initialized ==="
Write-Host " - Cosmos DB Data : '$env:COSMOSDB_DATA_PATH'"
Write-Host " - Sources        : '$env:SOURCES'"

Write-Host "=== Starting Testing Containers ==="
& docker-compose --compatibility build 2>&1
& docker-compose --compatibility `
    up `
    --exit-code-from tester `
    --always-recreate-deps `
    --abort-on-container-exit 2>&1
$result = $LASTEXITCODE

Write-Host "Cleaning Up"
& docker-compose --compatibility rm -s -f  2>&1

Write-Host "Exiting with $result"
exit $result