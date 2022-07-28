# Google vCard importer

This tool can be used to import contacts from vCards in _bulk_ into your Google Contacts accounts.

## Overview

This tool requests permission to authenticate with your Google account using OAuth 2.0. It can import a single or multiple vCard files into the account's contacts. It will create new contacts and applies a label `Imported vCard: <time>` to the newly created contacts to easily find them.

## Before you begin

To run this tool, you need the following prerequisites:
* [Go](https://golang.org/), latest version recommended.
* [Git](https://git-scm.com/), latest version recommended.
* A Google account.
* A Google Cloud Platform project with the API enabled. To create a project and enable an API, refer to [Create a project and enable the API](https://developers.google.com/workspace/guides/create-project).
* Authorization credentials for a desktop application. To learn how to create credentials for a desktop application, refer to [Create credentials](https://developers.google.com/workspace/guides/create-credentials).

## Run the tool to import vCards to Google Contacts

### Steps
1. Clone this repository.
2. Copy your Google dekstop application credentials to a file called `credentials.json` at the root level of the created workspace.
3. Run `go run main.go -f <path to your vCard(folder)>` from the workspace root. Replace `<path to your vCard(folder)>` with the path to either the vCard file or folder contains multiple vCard files that you want to import to Google.
    * _Optionally:_ Run `go run main.go` without additional argument for a demo run using the example vCard file `example.vcf`.

The first time you run the sample, it prompts you to authorize access:

4. Browse to the provided URL in your web browser.

If you're not already signed in to your Google account, you're prompted to sign in. If you're signed in to multiple Google accounts, you are asked to select one account to use for authorization.

5. Click the **Accept** button.
6. Copy the code you're given, paste it into the command-line prompt, and press **Enter**.

The tool will create contacts in your Google account and create a new label `Imported vCard: <time>` that is applied to each new contact for easy finding.

### Notes
- Authorization information is stored on the file system, so subsequent executions don't prompt for authorization.
- The authorization flow in this example is designed for a command-line application. For information on how to perform authorization in a web application, see [Using OAuth 2.0 for Web Server Applications](https://developers.google.com/accounts/docs/OAuth2WebServer).

## Troubleshooting

* For authentication issues see these [Google docs](https://developers.google.com/people/quickstart/go#troubleshooting) for common issues and solutions


## Next steps

* From within the Google Contacts application you can merge any newly created contacts with potential duplicated prior existing contacts [here](https://contacts.google.com/suggestions).