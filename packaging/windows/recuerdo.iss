; Inno Setup script for Recuerdo. Built by .github/workflows/release.yml:
;   iscc /DAppVersion=0.1.0 /DSourceDir=<deployed folder> packaging\windows\recuerdo.iss
; SourceDir holds recuerdo.exe with its Qt and MinGW DLLs, Qt plugins,
; data\ and config\.

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef SourceDir
  #define SourceDir "..\..\build\windows"
#endif

[Setup]
AppId={{9C2E4A71-5B3D-4F6A-8E21-4ECD0E0D0001}
AppName=Recuerdo
AppVersion={#AppVersion}
AppPublisher=Joop Kiefte
AppPublisherURL=https://github.com/LaPingvino/recuerdo
DefaultDirName={autopf}\Recuerdo
DefaultGroupName=Recuerdo
UninstallDisplayIcon={app}\recuerdo.exe
SetupIconFile=recuerdo.ico
OutputDir=..\..\dist
OutputBaseFilename=Recuerdo-{#AppVersion}-Setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
; install for the current user unless the user chooses all users
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "{#SourceDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Recuerdo"; Filename: "{app}\recuerdo.exe"
Name: "{group}\{cm:UninstallProgram,Recuerdo}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\Recuerdo"; Filename: "{app}\recuerdo.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\recuerdo.exe"; Description: "{cm:LaunchProgram,Recuerdo}"; Flags: nowait postinstall skipifsilent
