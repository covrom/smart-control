!include "MUI2.nsh"
!include "WinMessages.nsh"

;--------------------------------
; General

Name "SMARTDataCollector"
OutFile "SMARTDataCollectorInstaller.exe"
InstallDir $PROGRAMFILES\SMARTDataCollector
InstallDirRegKey HKCU "Software\SMARTDataCollector" "InstallDir"

;--------------------------------
; Interface Settings

!define MUI_ABORTWARNING

;--------------------------------
; Pages

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
Page custom MyConfigPageCreate MyConfigPageLeave
!insertmacro MUI_PAGE_INSTFILES

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

;--------------------------------
; Languages

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "Russian"

;--------------------------------
; Configuration Page

Var hCtl_HW
Var hCtl_APIUrl
Var hCtl_Token
Var hCtl_Cron

Function MyConfigPageCreate
    !insertmacro MUI_HEADER_TEXT "Configuration" "Enter service configuration parameters"
    nsDialogs::Create 1018
    Pop $0

    ${NSD_CreateLabel} 0 10u 100% 12u "Hostname:"
    Pop $0
    ${NSD_CreateText} 0 25u 100% 12u ""
    Pop $hCtl_HW
    ${NSD_SetText} $hCtl_HW "localhost"

    ${NSD_CreateLabel} 0 45u 100% 12u "API URL:"
    Pop $0
    ${NSD_CreateText} 0 60u 100% 12u ""
    Pop $hCtl_APIUrl
    ${NSD_SetText} $hCtl_APIUrl "http://192.168.169.12:18800/smart/report"

    ${NSD_CreateLabel} 0 80u 100% 12u "Token:"
    Pop $0
    ${NSD_CreateText} 0 95u 100% 12u ""
    Pop $hCtl_Token
    ${NSD_SetText} $hCtl_Token "EzEIhY8QtK6b"

    ${NSD_CreateLabel} 0 115u 100% 12u "Cron:"
    Pop $0
    ${NSD_CreateText} 0 130u 100% 12u ""
    Pop $hCtl_Cron
    ${NSD_SetText} $hCtl_Cron "55 23 * * *"

    nsDialogs::Show
FunctionEnd

Function MyConfigPageLeave
    ${NSD_GetText} $hCtl_HW $0
    StrCpy $R0 $0

    ${NSD_GetText} $hCtl_APIUrl $0
    StrCpy $R1 $0

    ${NSD_GetText} $hCtl_Token $0
    StrCpy $R2 $0

    ${NSD_GetText} $hCtl_Cron $0
    StrCpy $R3 $0

    ; Сохраняем в переменные, чтобы использовать позже
    WriteRegStr HKCU "Software\SMARTDataCollector" "Hostname" $R0
    WriteRegStr HKCU "Software\SMARTDataCollector" "ApiUrl" $R1
    WriteRegStr HKCU "Software\SMARTDataCollector" "Token" $R2
    WriteRegStr HKCU "Software\SMARTDataCollector" "Cron" $R3
FunctionEnd

;--------------------------------
; Installer Sections

Section "SMARTDataCollector" SecMain
    SectionIn RO

    SetOutPath "$INSTDIR"
    File "SMARTDataCollector.exe"  ; Замените на ваш исполняемый файл службы

    ; Создаем INI файл с параметрами
    WriteINIStr "$INSTDIR\settings.ini" "Config" "hostname" "$R0"
    WriteINIStr "$INSTDIR\settings.ini" "Config" "apiUrl" "$R1"
    WriteINIStr "$INSTDIR\settings.ini" "Config" "token" "$R2"
    WriteINIStr "$INSTDIR\settings.ini" "Config" "cron" "$R3"

    ; Установка службы
    ; ExecWait '"$INSTDIR\SMARTDataCollector.exe" install'
    ; Установка службы через sc create
    ExecWait '"$SYSDIR\sc.exe" create "SMARTDataCollector" binPath= "$INSTDIR\SMARTDataCollector.exe" start= auto DisplayName= "SMART Data Collector"' $0
    ; Запуск службы
    ExecWait '"$SYSDIR\sc.exe" start "SMARTDataCollector"' $0

    WriteRegStr HKCU "Software\SMARTDataCollector" "InstallDir" $INSTDIR
    WriteUninstaller "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Start Menu Shortcut"
    CreateDirectory "$SMPROGRAMS\SMARTDataCollector"
    CreateShortCut "$SMPROGRAMS\SMARTDataCollector\Uninstall.lnk" "$INSTDIR\Uninstall.exe"
SectionEnd

;--------------------------------
; Uninstaller Section

Section "Uninstall"
    ; Остановка службы перед удалением (если запущена)
    ExecWait '"$SYSDIR\sc.exe" stop "SMARTDataCollector"' $0
    Sleep 1000 ; Дадим время на остановку

    ; Удаление службы
    ; ExecWait '"$INSTDIR\SMARTDataCollector.exe" uninstall'
    ; Удаление службы через sc delete
    ExecWait '"$SYSDIR\sc.exe" delete "SMARTDataCollector"' $0

    Delete "$INSTDIR\settings.ini"
    Delete "$INSTDIR\SMARTDataCollector.exe"
    Delete "$INSTDIR\Uninstall.exe"
    RMDir "$INSTDIR"

    Delete "$SMPROGRAMS\SMARTDataCollector\Uninstall.lnk"
    RMDir "$SMPROGRAMS\SMARTDataCollector"

    DeleteRegKey /ifempty HKCU "Software\SMARTDataCollector"
SectionEnd