# WebNotes Backend

Welcome to the **WebNotes Backend** repository! This service acts as the backend api provider for the WebNotes Chrome Extension. It facilitates communication with the Notion API, handling OAuth flows, and managing web highlights to directly save them into users' Notion databases.

## 🚀 Features

- **Notion OAuth Integration**: Seamlessly connects users' Notion workspaces.
- **Highlight Management**: Save, retrieve, and delete highlights from web pages directly to/from Notion.
- **Database Provisioning**: Automatically ensure and create required database structures in Notion for storing highlights.
- **Cross-Origin Resource Sharing (CORS)**: Configured specifically for Chrome extension interactions.

## 🛠 Tech Stack

- **Go (Golang)**: Core programming language.
- **Gin**: High-performance HTTP web framework.
- **SQLite**: Local structured storage for access tokens and minimal state.
- **Notion API**: Primary external service for note synchronization.

---

## 🏎 Quick Start

Follow these steps to get the project up and running locally.

### 1. Prerequisites
- **Go** (v1.20+ recommended)
- **Git**
- A **Notion Integrations Account** (to obtain your Client ID and Secret, set up an integration at [Notion Developers](https://www.notion.so/my-integrations))

### 2. Clone the Repository
```bash
git clone https://github.com/Atoo35/WebNotes-backend.git
cd WebNotes-backend
```

### 3. Environment Variables
Create a `.env` file in the root directory. You can use the following template to set things up quickly:

```env
# Server
APP_ENV=dev
ALLOWED_ORIGINS=chrome-extension://<your-extension-id>
REDIRECT_URL=https://<your-extension-id>.chromiumapp.org

# Notion Integration
NOTION_CLIENT_ID=your_notion_client_id_here
NOTION_CLIENT_SECRET=your_notion_client_secret_here
NOTION_ROOT_URL=https://api.notion.com/v1
```
*Note: Make sure to replace placeholders with your actual extension ID and Notion credentials.*

### 4. Install Dependencies
```bash
go mod tidy
```

### 5. Run the Server
```bash
go run main.go
```
The server will start on `http://localhost:8080`.

---

## 📌 Main Endpoints (`/v1/notion`)

- `GET /start-oauth`: Initiates the Notion authentication flow.
- `GET /get-token`: Retrieves and stores the user token following OAuth.
- `POST /save-highlight`: Captures web highlights and pushes them to Notion.
- `GET /load-highlights-by-domain`: Fetches previously saved highlights for a specific website domain.
- `DELETE /delete-highlight`: Removes a stored highlight. 

---

## 🤝 Contributing
Contributions, issues, and feature requests are welcome! Feel free to check the issues page.

## 📝 License
This project is open-source and available for any modifications or redistributions.
