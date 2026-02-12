# 🎯 Job Tracker - Smart Job Application Manager

A browser extension that extracts job postings from any website, analyzes them using AI, and provides powerful analytics to help you track and understand job opportunities.

## ✨ Features

- **🔍 One-Click Job Extraction**: Extract job postings from any website with a single click
- **🤖 AI-Powered Analysis**: Automatically extract structured data (title, company, salary, skills, etc.) using Claude AI
- **📊 Auto-Generated Dashboards**: Beautiful Metabase dashboards created automatically on startup
- **💾 SQLite Database**: All your job data stored locally in a portable database
- **🎨 Skills Intelligence**: Track in-demand skills, programming languages, and cloud platforms
- **🔗 Skill Relationships**: Discover which skills commonly appear together
- **📈 Market Insights**: Analyze job market trends, salary ranges, and company types

## 🏗️ Architecture

Browser Extension (background.js) → Native Host (Go Binary) → Claude API + SQLite Database → Metabase Analytics

## 📋 Prerequisites

- **Go 1.25+** - Download from golang.org
- **Docker & Docker Compose** - Download from docker.com
- **Chrome or Firefox** browser
- **Anthropic API Key** - Get one at console.anthropic.com

## 🚀 Quick Start

### 1. Clone the Repository

git clone https://github.com/gypelayo/job-tool
cd job-tracker

### 2. Build the Native Host

cd native-host
go mod download
go build -o job-extractor cmd/main.go
chmod +x job-extractor

### 3. Install the Native Host Manifest

**For Linux/macOS Chrome:**
- Copy com.textextractor.host.json to ~/.config/google-chrome/NativeMessagingHosts/
- Update the path in the manifest to point to your job-extractor binary

**For Linux/macOS Firefox:**
- Copy com.textextractor.host.json to ~/.mozilla/native-messaging-hosts/
- Update the path in the manifest to point to your job-extractor binary

**For Windows:**
- Install via registry at HKEY_CURRENT_USER\Software\Google\Chrome\NativeMessagingHosts\com.textextractor.host

### 4. Install the Browser Extension

**Chrome:**
1. Open chrome://extensions/
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select the `extension` folder

**Firefox:**
1. Open about:debugging#/runtime/this-firefox
2. Click "Load Temporary Add-on"
3. Select `manifest.json` from the `extension` folder

### 5. Configure Your API Key

1. Click the extension icon in your browser toolbar
2. Right-click and select "Options" (or click the settings icon)
3. Enter your Anthropic API Key
4. Click "Save"

**🔐 Important**: Your API key is stored securely in the browser's storage, not in any files.

### 6. Start Metabase Analytics
```
cd native-host
docker compose up -d
docker compose logs -f metabase-setup
```
Once complete, open http://localhost:3000

**Login Credentials:**
- Email: admin@example.com
- Password: SecurePassword123!

## 📖 Usage

### Extracting a Job Posting

1. Navigate to any job posting webpage
2. Click the Job Tracker extension icon in your browser
3. Wait for the AI to analyze the page
4. View the extracted data in the popup

### Viewing Analytics

1. Open http://localhost:3000
2. Navigate to "Skills Intelligence Dashboard"
3. Explore:
   - Top 20 most in-demand skills
   - Skills by category breakdown
   - Programming language trends
   - Cloud platform demand
   - Common skill combinations

## 🗄️ Database Schema

**Jobs Table:**
- id, job_title, company_name, location
- salary_min, salary_max, employment_type
- remote_friendly, seniority_level, etc.

**Job Skills Table:**
- id, job_id, skill_name, skill_category
- Foreign key relationship to jobs table

## 🔧 Configuration

### API Key Configuration

The API key is configured through the extension's Options page:
1. Right-click the extension icon → Options
2. Enter your Anthropic API key
3. The key is passed to the native host with each extraction request

### Database Location

By default, the database is stored at: ~/Downloads/extracted_jobs/jobs.db

You can change this in the native-host configuration.

### Docker Compose

Edit `docker-compose.yml` to customize:
- Metabase port (default: 3000)
- Database volume path
- Admin credentials (in cmd/metabase-setup/main.go)

## 🛠️ Development

### Running Tests
```
cd native-host && go test ./...
```
### Debugging

Test the native host directly by piping JSON to it or check browser console for logs.

### Adding New Visualizations

Edit `cmd/metabase-setup/main.go` and add new cards to the slice, then rebuild with docker compose.

## 🐛 Troubleshooting

### Extension Not Connecting

1. Verify native host manifest path is correct
2. Check binary is executable
3. Review browser console for errors
4. Check extractor.log in native-host directory

### API Key Issues

- Make sure you've entered the API key in the extension's Options page
- Verify the key is valid at console.anthropic.com
- Check browser console for authentication errors

### Database Not Found

- Create ~/Downloads/extracted_jobs directory
- Set proper permissions (755)
- Check extractor.log for database errors

### Metabase Dashboard Empty

- Restart with docker compose down -v && docker compose up -d
- Verify database has data with sqlite3 queries
- Check that database path in docker-compose.yml matches actual location

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push and open a Pull Request

## 📝 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- Metabase for analytics platform
- Go for native host implementation

## 📧 Support

Open an issue on GitHub for help

---

**Made with ❤️ for job seekers everywhere**
