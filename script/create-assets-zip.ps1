# PowerShell script to create a compressed ZIP file of executable assets
# This script compresses all executable files from the assets folder into installers.zip

param(
    [string]$AssetsPath = "assets",
    [string]$OutputZip = "installers.zip"  # Changed to root directory
)

# Verify assets folder exists
if (-not (Test-Path $AssetsPath)) {
    Write-Error "Assets folder not found: $AssetsPath"
    exit 1
}

# Remove existing zip if present
if (Test-Path $OutputZip) {
    Remove-Item $OutputZip -Force
    Write-Host "Removed existing zip: $OutputZip"
}

$assetsFullPath = (Get-Item $AssetsPath).FullName
$zipPath = [System.IO.Path]::GetFullPath($OutputZip)

try {
    # Get all asset files (exclude only .yaml config files)
    # Include .txt, .exe, .ini and all other files
    $files = @(Get-ChildItem -Path $assetsFullPath -File | Where-Object {
        $_.Extension -ne ".yaml"
    })
    
    if ($files.Count -eq 0) {
        Write-Error "No asset files found in assets directory"
        exit 1
    }
    
    # Get original size
    $originalSize = ($files | Measure-Object -Property Length -Sum).Sum
    
    # Create zip using Compress-Archive (PowerShell 5.0+)
    Write-Host "Compressing $($files.Count) asset file(s) to $OutputZip..."
    $filesToZip = $files | Select-Object -ExpandProperty FullName
    Compress-Archive -Path $filesToZip -DestinationPath $zipPath -Force -CompressionLevel Optimal
    
    # Get compressed size
    $compressedSize = (Get-Item $zipPath).Length
    $compressionRatio = if ($originalSize -gt 0) { [math]::Round((1 - $compressedSize / $originalSize) * 100, 2) } else { 0 }
    
    Write-Host "`nZip created successfully: $OutputZip"
    Write-Host "Files compressed: $($files.Count)"
    Write-Host "Files included:"
    $files | ForEach-Object { Write-Host "  - $($_.Name)" }
    Write-Host "`nOriginal size: $([math]::Round($originalSize / 1MB, 2)) MB"
    Write-Host "Compressed size: $([math]::Round($compressedSize / 1MB, 2)) MB"
    Write-Host "Compression ratio: $compressionRatio%"
}
catch {
    Write-Error "Failed to create zip: $_"
    exit 1
}
