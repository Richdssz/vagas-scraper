@echo off
chcp 65001 > nul
echo ====================================================
echo   🚀 Enviando alterações para o GitHub...
echo ====================================================
echo.

git add .
git commit -m "chore: atualizar arquivos locais"
git push origin main

echo.
echo ====================================================
echo   ✅ Tudo atualizado no GitHub com sucesso!
echo ====================================================
pause
