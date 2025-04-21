# Rail Go Bot

A Telegram bot for checking Israel Railways schedules.

## Features

- Check train schedules between stations
- Predefined routes (home/work)
- Custom route selection
- Monthly notifications
- Caching of schedule results

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/rail-go.git
cd rail-go
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Edit the `.env` file with your configuration:
- `BOT_TOKEN`: Your Telegram bot token from @BotFather
- `TRAIN_API_KEY`: Your Israel Railways API key
- Other settings as needed

4. Build and run:
```bash
go build
./rail-go
```

## Configuration

The application can be configured through environment variables:

### Bot Configuration
- `BOT_TOKEN`: Telegram bot token
- `BOT_DEBUG`: Enable debug mode (true/false)
- `BOT_TIMEOUT`: Update timeout in seconds

### Scheduler Configuration
- `SCHEDULER_RETRY_ATTEMPTS`: Number of retry attempts for failed tasks

### Notifications Configuration
- `DEFAULT_CHAT_ID`: Default chat ID for notifications
- `TIME_FORMAT`: Time format for notifications
- `MONTHLY_TEMPLATE`: Template for monthly notifications

### Train Service Configuration
- `TRAIN_API_KEY`: Israel Railways API key
- `TRAIN_USER_AGENT`: User agent for API requests
- `TRAIN_BASE_URL`: Base URL for API requests
- `TRAIN_TIMEOUT`: Request timeout in seconds

## Usage

1. Start a chat with your bot
2. Use the following commands:
   - `/start` - Start the bot
   - `/help` - Show help message
   - `/home` - Check schedule for home route
   - `/work` - Check schedule for work route
   - `/other` - Select custom route

## Development

### Project Structure

```
rail-go/
├── cmd/
│   └── bot/
│       └── main.go
├── internal/
│   ├── bot/
│   │   ├── handler.go
│   │   └── service.go
│   ├── cache/
│   │   └── cache.go
│   ├── config/
│   │   └── config.go
│   └── train/
│       ├── client.go
│       └── models.go
├── pkg/
│   └── logger/
│       └── logger.go
├── .env
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

### Adding New Features

1. Create a new branch for your feature
2. Implement the feature
3. Add tests
4. Update documentation
5. Create a pull request

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 