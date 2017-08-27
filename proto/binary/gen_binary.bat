set CURR=%cd%
cd ..\..\..\..\..\..
set GOPATH=%cd%
cd %CURR%

go build -o objprotogen.exe github.com/mutousay/cellnet/objprotogen
@IF %ERRORLEVEL% NEQ 0 pause


objprotogen.exe --out coredef/objproto_gen.go coredef/binmsg.go
@IF %ERRORLEVEL% NEQ 0 pause
