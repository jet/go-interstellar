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

$sleeptime = 5

Function Import-CosmosDBCert {
    $importcertScript = "${env:COSMOSDB_DATA_PATH}\importcert.ps1"
    if( Test-Path $importcertScript ) {
        Write-Host "Importing Certificates"
        & $importcertScript
        return $True
    }
    return $False
}

Function Get-CosmosDBStatusReady {
    try {
        Invoke-RestMethod -Uri "https://localhost:8081/" | Out-Null
    } catch [System.Net.WebException] {
        $status = [int]($_.Exception.Response.StatusCode)
        if ($status -eq 401) { # Unauthorized means it's up!
            return $True
        }
        Write-Host "CosmosDB Response: $status"
        $msg = $_.Exception.Message
        Write-Host "Exception: $msg"
        return $False
    } catch {
        return $False
    }
    return $True
}

$count = 0
Write-Host "Waiting for CosmosDB Certificate..."
While (-not (Import-CosmosDBCert)) {
    $count = $count + $sleeptime
    if($count -gt 1000) {
        Write-Host "Timed Out"
        exit 1
    }
    Write-Host "CosmosDB Certificate missing; Sleeping"
    Start-Sleep -Seconds $sleeptime

}
Write-Host "CosmosDB Certificate installed!"

$count = 0
Write-Host "Waiting for CosmosDB to Start..."
While (-not (Get-CosmosDBStatusReady)) {
    $count = $count + $sleeptime
    if($count -gt 1000) {
        Write-Host "Timed Out"
        exit 1
    }
    Write-Host "CosmosDB Not Ready; Sleeping"
    Start-Sleep -Seconds $sleeptime
}
Write-Host "CosmosDB is Ready!"

Write-Host "Running Integration Tests"
$env:DEBUG_LOGGING="Y"
$env:RUN_INTEGRATION_TESTS="Y"
$env:AZURE_COSMOS_DB_CONNECTION_STRING="AccountEndpoint=https://localhost:8081/;AccountKey=C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
Set-Location -Path "$env:BUILD_SOURCESDIRECTORY"
& go test -v -cover .
$result = $LASTEXITCODE
Write-Host "Exiting with $result"
exit $result