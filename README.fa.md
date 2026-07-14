# RedScope

RedScope یک مانیتور زنده شبکه داخل ترمینال است.

نشان می‌دهد هر برنامه روی سیستم به کدام IP یا hostname وصل شده است.

![جای اسکرین‌شات](docs/screenshot.png)

## قابلیت‌ها

- رابط TUI با Go و [`tview`](https://github.com/rivo/tview)
- نمایش نام پردازش و PID
- لیست اتصال‌های TCP/UDP
- آدرس و پورت local و remote
- پیدا کردن hostname با reverse DNS
- نمایش وضعیت اتصال
- فیلتر و جستجو داخل TUI
- خروجی به صورت یک فایل اجرایی ساده

## نصب و اجرا

```bash
git clone https://github.com/TechnoRed2026/redscope.git
cd redscope
go mod tidy
go run .
```

ساخت فایل اجرایی:

```bash
go build -o redscope .
```

در ویندوز:

```powershell
go build -o redscope.exe .\
.\redscope.exe
```

## کلیدها

| کلید | عملکرد |
| --- | --- |
| `/` | فوکوس روی فیلتر |
| `Esc` | پاک کردن فیلتر / برگشت به جدول |
| `r` | رفرش دستی |
| `q` | خروج |

## نمونه خروجی

```text
Process       PID    Proto  Local              Remote             Host                 State
chrome.exe    4250   TCP    192.168.1.5:51231  140.82.113.6:443   lb-140-82-113-6...  ESTABLISHED
Code.exe      6112   TCP    192.168.1.5:51302  13.107.42.18:443   vscode-sync...      ESTABLISHED
```

## محدودیت‌ها

RedScope فقط metadata شبکه را نشان می‌دهد. محتوای HTTPS یا packetها را decrypt نمی‌کند.

در بعضی سیستم‌ها برای دیدن همه پردازش‌ها باید برنامه را با Administrator/root اجرا کنید.

## نقشه راه

- نمایش مصرف پهنای باند برای هر process
- تاریخچه DNS queryها
- GeoIP / ASN
- هشدار برای مقصدهای مشکوک
- خروجی CSV/JSON

## لایسنس

MIT
