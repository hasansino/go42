# Setting Up Gemini GitHub App

## Step 1: Create the GitHub App

1. Go to your GitHub Settings:
   - Personal: https://github.com/settings/apps/new
   - Organization: https://github.com/organizations/[YOUR_ORG]/settings/apps/new

2. Fill in the App details:
   - **Name**: `Gemini AI Assistant` (or your preference)
   - **Homepage URL**: `https://github.com/hasansino/go42`
   - **Description**: `AI-powered assistant using Google Gemini`
   - **Webhook**: Uncheck "Active" (not needed)

3. Set Permissions:
   - **Repository permissions**:
     - Contents: Read & Write
     - Issues: Read & Write
     - Pull requests: Read & Write
     - Metadata: Read
   - **Organization permissions**: None needed

4. Where can this GitHub App be installed?
   - Choose "Only on this account"

5. Click "Create GitHub App"

## Step 2: Configure the App

1. After creation, you'll see your App's page
2. Note down the **App ID** (you'll need this)
3. Scroll down to "Private keys"
4. Click "Generate a private key"
5. Save the downloaded `.pem` file

## Step 3: Install the App

1. On your App's page, click "Install App"
2. Choose your repository or organization
3. Select "Only select repositories" 
4. Choose `hasansino/go42`
5. Click "Install"

## Step 4: Configure Repository

Add these as repository secrets and variables:

### Repository Variables (Settings → Secrets → Variables → New repository variable):
```
APP_ID=<your-app-id>
```

### Repository Secrets (Settings → Secrets → Actions → New repository secret):
```
APP_PRIVATE_KEY=<contents-of-your-pem-file>
```

## Step 5: Update Workflow (Optional)

Your workflow already has the GitHub App support built in! The workflows will automatically use the App token when `APP_ID` is set.

## Result

Now when Gemini responds, comments will show as:
- **Author**: "Gemini AI Assistant [bot]" (with your custom avatar)
- **Clear attribution**: Users can easily see it's from Gemini
- **Professional appearance**: Looks like other GitHub integrations

## Adding a Custom Avatar

1. Go to your App's settings
2. Upload a logo (recommended: 256x256px PNG)
3. Suggested: Use a Gemini-themed icon or AI-related image

## Benefits

- ✅ Clear attribution - no more generic "github-actions"
- ✅ Custom avatar makes it visually distinctive
- ✅ Professional appearance
- ✅ Better rate limits than GITHUB_TOKEN
- ✅ Can be managed separately from repository permissions