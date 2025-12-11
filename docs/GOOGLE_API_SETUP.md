# Google Calendar API Setup Guide

This guide walks you through setting up Google Calendar API credentials for Tempus, step by step.

## Overview

Tempus uses OAuth 2.0 Device Flow to authenticate with Google Calendar. You'll need:

1. **Google Cloud Project** (free)
2. **OAuth 2.0 Client ID and Secret** (one-time setup)
3. **Google Calendar API enabled**

**Time required**: ~10 minutes for first-time setup

---

## Step 1: Create a Google Cloud Project

### 1.1 Go to Google Cloud Console

Visit [https://console.cloud.google.com/](https://console.cloud.google.com/)

- Sign in with your Google account
- You may need to accept Terms of Service on first visit

### 1.2 Create a New Project

1. Click the project dropdown in the top navigation (says "Select a project")
2. Click **"NEW PROJECT"** in the top right
3. Fill in the project details:
   - **Project name**: `Tempus Calendar` (or any name you prefer)
   - **Organization**: Leave as "No organization" (for personal use)
4. Click **"CREATE"**

Wait ~30 seconds for the project to be created, then select it from the dropdown.

---

## Step 2: Enable Google Calendar API

### 2.1 Open the API Library

1. In the left sidebar, click **"APIs & Services"** → **"Library"**
   - Or search "API Library" in the top search bar
2. In the search box, type: `Google Calendar API`
3. Click on **"Google Calendar API"** from the results

### 2.2 Enable the API

1. Click the blue **"ENABLE"** button
2. Wait for the API to be enabled (~10 seconds)
3. You'll be redirected to the API dashboard

---

## Step 3: Configure OAuth Consent Screen

### 3.1 Navigate to OAuth Consent

1. In the left sidebar, click **"OAuth consent screen"**
2. Choose **"External"** user type (for personal use)
   - This allows you to use the app yourself without organization restrictions
3. Click **"CREATE"**

### 3.2 Fill in App Information

**App information:**
- **App name**: `Tempus` (or any name you prefer)
- **User support email**: Your email address (select from dropdown)
- **App logo**: Optional - skip for personal use

**App domain:** (Optional - can leave blank for personal use)
- Application home page: (leave blank)
- Application privacy policy: (leave blank)
- Application terms of service: (leave blank)

**Developer contact information:**
- **Email addresses**: Your email address

Click **"SAVE AND CONTINUE"**

### 3.3 Add Scopes

1. Click **"ADD OR REMOVE SCOPES"**
2. In the filter box, search for: `calendar`
3. Check the box for:
   - ✅ `https://www.googleapis.com/auth/calendar` (Full access to Google Calendar)
4. Click **"UPDATE"** at the bottom
5. Click **"SAVE AND CONTINUE"**

### 3.4 Add Test Users (Required for External Apps)

1. Click **"ADD USERS"**
2. Enter your Gmail address (the one you'll use with Tempus)
3. Click **"ADD"**
4. Click **"SAVE AND CONTINUE"**

### 3.5 Review and Submit

1. Review the summary
2. Click **"BACK TO DASHBOARD"**

**Note**: Your app will stay in "Testing" mode, which is fine for personal use. It allows up to 100 test users.

---

## Step 4: Create OAuth 2.0 Credentials

### 4.1 Navigate to Credentials

1. In the left sidebar, click **"Credentials"**
2. Click **"+ CREATE CREDENTIALS"** at the top
3. Select **"OAuth client ID"**

### 4.2 Configure OAuth Client

1. **Application type**: Select **"Desktop app"**
   - This is important! Tempus uses Device Flow, which requires Desktop app type
2. **Name**: `Tempus CLI` (or any name you prefer)
3. Click **"CREATE"**

### 4.3 Copy Your Credentials

A popup will appear showing your credentials:

- **Client ID**: Something like `123456789-abcdefghijklmnop.apps.googleusercontent.com`
- **Client secret**: Something like `GOCSPX-AbCdEfGhIjKlMnOpQrStUvWx`

**Important**:
- Click **"DOWNLOAD JSON"** to save the credentials file (optional backup)
- Copy both values - you'll need them in the next step

---

## Step 5: Configure Tempus

### Option 1: Environment Variables (Recommended)

Set environment variables in your shell:

```bash
# Add to your ~/.bashrc, ~/.zshrc, or ~/.profile
export TEMPUS_CLIENT_ID="YOUR-CLIENT-ID.apps.googleusercontent.com"
export TEMPUS_CLIENT_SECRET="YOUR-CLIENT-SECRET"
```

Then reload your shell:
```bash
source ~/.bashrc  # or ~/.zshrc
```

### Option 2: .env File (Alternative)

Create a `.env` file in your home directory or project folder:

```bash
echo 'TEMPUS_CLIENT_ID="YOUR-CLIENT-ID.apps.googleusercontent.com"' >> ~/.tempus.env
echo 'TEMPUS_CLIENT_SECRET="YOUR-CLIENT-SECRET"' >> ~/.tempus.env

# Load before running Tempus
source ~/.tempus.env
```

### Option 3: Command-Line Flags (Least Secure)

Pass credentials as flags (not recommended - visible in process list):

```bash
tempus google import \
  --client-id "YOUR-CLIENT-ID.apps.googleusercontent.com" \
  --client-secret "YOUR-CLIENT-SECRET" \
  --input event.ics \
  --calendar primary \
  --token-file ~/.tempus/google_token.json
```

---

## Step 6: First Authorization

### 6.1 Import an ICS File

```bash
# Create a test ICS file first
tempus create \
  --summary "Test Event" \
  --start "2025-11-15 10:00" \
  --end "2025-11-15 11:00" \
  --output test.ics

# Import to Google Calendar
tempus google import \
  --input test.ics \
  --calendar primary \
  --token-file ~/.tempus/google_token.json
```

### 6.2 Complete Device Authorization

Tempus will display:

```
Authorize Tempus at https://www.google.com/device with code: ABCD-EFGH
Waiting for authorization...
```

**Steps:**

1. Open the URL in your browser: `https://www.google.com/device`
2. Enter the code shown (e.g., `ABCD-EFGH`)
3. Click **"Next"**
4. **Choose your Google account**
5. Review the permissions:
   - "Tempus wants to access your Google Account"
   - "See, edit, share, and permanently delete all calendars you can access"
6. Click **"Continue"**
7. You'll see: "You've given Tempus permission to access your data"

### 6.3 Token Storage

After authorization:
- Tempus saves the token to `~/.tempus/google_token.json` (or your specified path)
- Future imports will reuse this token automatically
- Tokens are valid for ~7 days and auto-refresh

**Security note**: Protect your token file:
```bash
chmod 600 ~/.tempus/google_token.json
```

---

## Step 7: Verify Setup

### Test Importing an Event

```bash
# Create an event
tempus create \
  --summary "Dentist Appointment" \
  --start "2025-11-20 14:00" \
  --duration 1h \
  --output dentist.ics

# Import to Google Calendar
tempus google import \
  --input dentist.ics \
  --calendar primary \
  --token-file ~/.tempus/google_token.json
```

**Expected output**:
```
✅ Imported ICS into Google Calendar primary
```

Check your Google Calendar - the event should appear!

---

## Common Issues

### Issue: "Access blocked: Authorization Error"

**Problem**: Your app is in Testing mode and your email isn't added as a test user.

**Solution**:
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **"APIs & Services"** → **"OAuth consent screen"**
3. Scroll to **"Test users"** section
4. Click **"ADD USERS"** and add your Gmail address

---

### Issue: "Invalid client: Unauthorized"

**Problem**: Client ID or Secret is incorrect.

**Solution**:
1. Verify your credentials in [Google Cloud Console](https://console.cloud.google.com/)
2. Go to **"APIs & Services"** → **"Credentials"**
3. Find your OAuth 2.0 Client ID and re-copy the values
4. Update your environment variables or flags

---

### Issue: "API has not been enabled"

**Problem**: Google Calendar API is not enabled for your project.

**Solution**:
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **"APIs & Services"** → **"Library"**
3. Search for "Google Calendar API"
4. Click **"ENABLE"**

---

### Issue: Token expired

**Problem**: Token file is old or invalid.

**Solution**:
```bash
# Delete old token
rm ~/.tempus/google_token.json

# Re-authorize (will prompt for device code again)
tempus google import --input event.ics --calendar primary --token-file ~/.tempus/google_token.json
```

---

### Issue: "Calendar not found"

**Problem**: Calendar ID is incorrect or you don't have access.

**Solution**:
- Use `primary` for your main calendar
- Or find the Calendar ID:
  1. Open [Google Calendar](https://calendar.google.com/)
  2. Click the 3 dots next to a calendar → "Settings and sharing"
  3. Scroll to "Integrate calendar" → copy the **Calendar ID**
  4. Use that ID: `--calendar "YOUR_CALENDAR_ID@group.calendar.google.com"`

---

## Advanced Usage

### Import to Specific Calendar

```bash
# List your calendars (if you have a list command, otherwise use web UI)
# Find the Calendar ID from Google Calendar settings

tempus google import \
  --input work-meeting.ics \
  --calendar "work-calendar-id@group.calendar.google.com" \
  --token-file ~/.tempus/google_token.json
```

### Batch Import Multiple Events

```bash
# Create events from CSV
tempus batch \
  --input events.csv \
  --template meeting \
  --output batch.ics

# Import all events at once
tempus google import \
  --input batch.ics \
  --calendar primary \
  --token-file ~/.tempus/google_token.json
```

### Use Different Token Files for Multiple Accounts

```bash
# Personal calendar
tempus google import \
  --input personal.ics \
  --calendar primary \
  --token-file ~/.tempus/personal_token.json

# Work calendar (will prompt for different account)
tempus google import \
  --input work.ics \
  --calendar primary \
  --token-file ~/.tempus/work_token.json
```

---

## Security Best Practices

### ✅ DO:
- Store credentials in environment variables
- Use `chmod 600` on token files
- Add `*_token.json` to `.gitignore`
- Revoke access if compromised (Google Account settings)
- Use separate tokens for different Google accounts

### ❌ DON'T:
- Commit credentials or tokens to version control
- Share your Client Secret publicly
- Use `sudo` to run Tempus (unnecessary)
- Paste credentials in public forums/chat

### Revoking Access

If you need to revoke Tempus access:

1. Go to [myaccount.google.com/permissions](https://myaccount.google.com/permissions)
2. Find "Tempus" (or your app name)
3. Click **"Remove Access"**
4. Delete your local token file: `rm ~/.tempus/google_token.json`

---

## FAQ

**Q: Do I need to publish my app?**
A: No! Testing mode is sufficient for personal use (up to 100 users).

**Q: Why does Google show a security warning?**
A: Because the app is unverified. This is normal for personal OAuth apps. Click "Advanced" → "Go to Tempus (unsafe)" to proceed.

**Q: Can I use this with Google Workspace (work account)?**
A: Yes, but your workspace admin may need to approve the app. Check with IT.

**Q: How long do tokens last?**
A: Access tokens expire after ~1 hour, but Tempus automatically refreshes them using the refresh token (valid for months).

**Q: Can I share my credentials with others?**
A: No - each user should create their own OAuth credentials for security.

**Q: Is this free?**
A: Yes! Google Calendar API has a generous free tier (1,000,000 requests/day).

---

## Getting Help

If you encounter issues not covered here:

1. Check the [SECURITY.md](../SECURITY.md) for security-related questions
2. Open an issue on GitHub with:
   - Your OS and Go version
   - The exact command you ran
   - The error message (redact any credentials!)
3. Join discussions in the repository

---

**Last updated**: 2025-11-13
**Tempus version**: 0.5.x+
