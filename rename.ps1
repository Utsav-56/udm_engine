$target_dir = "D:\GO_projects\nudm\.examples"

# Rename each and every .go files with extension .go.old simply append the suffix .old
Get-ChildItem -Path $target_dir -Filter *.go | ForEach-Object {
    Rename-Item -Path $_.FullName -NewName ($_.Name + ".old")
    Write-Host "Renamed $($_.Name) to $($_.Name + '.old')"
}
