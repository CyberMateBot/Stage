# Apply SQL migrations to local Postgres (docker compose service tgapp_postgres).
param(
    [string]$Container = "tgapp_postgres",
    [string]$Database = "myapp_db",
    [string]$User = "postgres"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$migrations = @(
    "V20250805000000__users.sql",
    "V20251103000100__cybermate_core.sql",
    "V20251103000200__admin_resources.sql",
    "V20250525000000__profile_ui_theme.sql"
)

foreach ($f in $migrations) {
    $path = Join-Path $root "internal\migrations\$f"
    Write-Host "Applying $f ..."
    Get-Content $path -Raw | docker exec -i $Container psql -U $User -d $Database -v ON_ERROR_STOP=1
}

Write-Host "Done. Tables:"
docker exec $Container psql -U $User -d $Database -c "\dt"
