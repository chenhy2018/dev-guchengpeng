@echo off
if "%1" == "" goto help
8g -o %1.8 %1
8l -o %1.exe %1.8
del %1.8
%1.exe %2 %3 %4 %5 %6 %7 %8 %9
goto done
:help
echo Usage: go $GoScriptFile
:done
