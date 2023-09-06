# Recipe Gin API

![License](https://img.shields.io/badge/License-MIT-blue.svg)

> A RESTful API built with Gin to manage recipes.

The Recipe Gin API is a simple RESTful API that allows you to manage recipes. It provides endpoints to create, read, update, and delete recipes, making it easy to organize and maintain your collection of recipes.

## Table of Contents

- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [License](#license)

## Features

- Create, read, update, and delete recipes.
- RESTful API design for easy integration.
- [`PENDING`] Authentication and authorization for secure access.
- Data validation and error handling.
- Lightweight and built with [Gin](https://github.com/gin-gonic/gin).

## Getting Started

### Prerequisites

Before you begin, ensure you have met the following requirements if you are developing/running locally:

- Go 1.20.5 or higher installed.
- Git installed.
- A NoSQL database (MongoDB) for storing recipe data.

Else, the following [installation](#installation) section provides instructions to run the application in a docker environment.

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/wtlow003/recipe-gin-api.git
   ```
2. Change to project directory:

   ```bash
   cd recipe-gin-api
   ```
3. Create a `.env` file based on the `.env.example` and configure it with your database settings and any other environment variables needed.
4. Run docker containers:

   ```bash
   docker compose up --build
   ```
   `NOTE`: Using `docker-compose` instead of `docker compose` may resulting in port conflict when setting replicas for api container.

The API should now be running on localhost (e.g., http://locahost:8080/api/v1/recipes) on port `8079-8081`.

## Usage

To use the Recipe Gin API, you can make HTTP requests to its endpoints. The API documentation provides details on available endpoints and how to interact with them.

Example:

1. Retrieve all available recipes

    ```bash
    curl http://localhost:8070/api/v1/recipes | jq -r

    >>>
    [
        {
            "id": "64d236d01af83c4f1209cdcf",
            "name": "Baked Shrimp Scampi",
            "tags": [
            "seafood",
            "shrimp",
            "main"
            ],
            "ingredients": [
            "2/3 cup panko\r",
            "1/4 teaspoon red pepper flakes\r",
            "1/2 lemon, zested and juiced\r",
            "1 extra-large egg yolk\r",
            "1 teaspoon rosemary, minced\r",
            "3 tablespoon parsley, minced\r",
            "4 clove garlic, minced\r",
            "1/4 cup shallots, minced\r",
            "8 tablespoon unsalted butter, softened at room temperature\r",
            "<hr>",
            "2 tablespoon dry white wine\r",
            "Freshly ground black pepper\r",
            "Kosher salt\r",
            "3 tablespoon olive oil\r",
            "2 pound frozen shrimp"
            ],
            "instructions": "Preheat the oven to 425 degrees F.\r\n\r\nDefrost shrimp by putting in cold water, then drain and toss with wine, oil, salt, and pepper. Place in oven-safe dish and allow to sit at room temperature while you make the butter and garlic mixture.\r\n\r\nIn a small bowl, mash the softened butter with the rest of the ingredients and some salt and pepper.\r\n\r\nSpread the butter mixture evenly over the shrimp. Bake for 10 to 12 minutes until hot and bubbly. If you like the top browned, place under a broiler for 1-3 minutes (keep an eye on it). Serve with lemon wedges and French bread.\r\n\r\nNote: if using fresh shrimp, arrange for presentation. Starting from the outer edge of a 14-inch oval gratin dish, arrange the shrimp in a single layer cut side down with the tails curling up and towards the center of the dish. Pour the remaining marinade over the shrimp. ",
            "servings": 6,
            "calories": 2565,
            "fat": 159,
            "satfat": 67,
            "carbs": 76,
            "fiber": 4,
            "sugar": 6,
            "protein": 200,
            "publishedAt": "0001-01-01T00:00:00Z"
        },
        ...
    ]
    ```

## API Documentation.

For detailed information on how to use the API, refer to documentation availble on [Swagger UI](http://localhost:8080/swagger/index.html).

## License
This project is licensed under the MIT License. See the '
[LICENSE](./LICENSE) file for details.