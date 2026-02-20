# 🎯 Job Tracker - Smart Job Application Manager


``` sh
mkdir -p ~/.mozilla/native-messaging-hosts
```

Change native-host/com.textextractor.firefox.json <project-folder> to this project's absolute path

```sh
cp com.textextractor.firefox.json ~/.mozilla/native-messaging-hosts/
chmod +x job-extractor
```


## Load the extension in Firefox (dev/test)

    1. Open Firefox.

    2. Go to about:debugging#/runtime/this-firefox.

    3. Click “Load Temporary Add-on…”.

    4. Select text-extractor/extension/manifest.json.

## Set Perplexity key in the extension

    1. Click the extension icon → open Dashboard.

    2. Go to Settings tab.

    3. Paste your Perplexity API key in “API Key”, click Save.
   
## Still in Settings tab:

    1. Click Test connection.

    2. Status should change to “Native helper is connected.”

## Use it: extract a job and see charts

    1. Open a job posting page in Firefox.

    2. Click the extension popup, hit Extract.

    3. You should see “Job extracted and saved.”

    4. Open the Dashboard (from popup or extension icon):

        - Jobs tab: job list, view details, set status, notes.

        - Analytics tab: charts by skill type, job titles, skills per stage.
