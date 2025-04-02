# DeFi Aggregator API Documentation

## Overview

The DeFi Aggregator is a service that allows users to view and compare DeFi projects and find the best trading routes across multiple decentralized exchanges. This document provides details on the available API endpoints and their usage.

## Base URL

```
https://api.gwei.exchange
```

## Authentication

Some endpoints require authentication using an API key. Include the key in the header of your requests:

```
x-api-key: your-api-key
```

## Endpoints

### Health Check

Check if the API is running properly.

**Endpoint:** `GET /`

**Response:**

```json
{
  "message": "Hello, World!"
}
```

### Get Supported Protocols

Retrieve a list of all supported DeFi protocols.

**Endpoint:** `GET /protocols`

**Response:**

```json
{
  "protocols": [
    "uniswapv3",
    "sushiswapv3",
    "pancakeswapv3",
    "tayaswap"
  ]
}
```

### Get Token Information

Retrieve metadata for a specific token by its address.

**Endpoint:** `GET /token`

**Query Parameters:**
- `address`: The Ethereum address of the token (required)

**Response:**

```json
{
  "source": "blockchain",
  "token": {
    "address": "0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D",
    "name": "USD Coin",
    "symbol": "USDC",
    "decimals": 6
  }
}
```

If the token data is found in cache:

```json
{
  "source": "cache",
  "token": {
    "address": "0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D",
    "name": "USD Coin",
    "symbol": "USDC",
    "decimals": 6
  }
}
```

### Save Token Information

Add or update token metadata in the system.

**Endpoint:** `POST /token`

**Headers:**
- `x-api-key`: Your API key (required)
- `Content-Type`: application/json

**Request Body:**

```json
{
  "address": "0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D"
}
```

**Response:**

```json
{
  "message": "Token metadata fetched and saved",
  "token": {
    "address": "0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D",
    "name": "USD Coin",
    "symbol": "USDC",
    "decimals": 6
  }
}
```

### Get Trading Pairs

Find optimal trading routes for a token pair.

**Endpoint:** `GET /pairs`

**Query Parameters:**
- `tokena`: The Ethereum address of the first token (required)
- `tokenb`: The Ethereum address of the second token (required)
- `amount`: The amount of tokena to swap (optional, default: 10000)
- `all`: Return all available routes (optional, default: false)

**Response:**

When `all=false` (default):

```json
{
  "result": {
    "protocol": "Uniswap V3",
    "poolAddress": "0x7F86Bf177Dd4F3494b841a37e810A34dD56c829B",
    "fee": 3000,
    "tokenIn": "USDC",
    "tokenOut": "WETH",
    "amountIn": "100",
    "amountOut": "0.05231"
  }
}
```

When `all=true`:

```json
{
  "bestRoute": {
    "protocol": "Uniswap V3",
    "poolAddress": "0x7F86Bf177Dd4F3494b841a37e810A34dD56c829B",
    "fee": 3000,
    "tokenIn": "USDC",
    "tokenOut": "WETH",
    "amountIn": "100",
    "amountOut": "0.05231"
  },
  "allRoutes": [
    {
      "protocol": "Uniswap V3",
      "poolAddress": "0x7F86Bf177Dd4F3494b841a37e810A34dD56c829B",
      "fee": 3000,
      "tokenIn": "USDC",
      "tokenOut": "WETH",
      "amountIn": "100",
      "amountOut": "0.05231"
    },
    {
      "protocol": "SushiSwap V3",
      "poolAddress": "0xD9E1cE17f2641f24aE83637ab66a2cca9C378B9F",
      "fee": 3000,
      "tokenIn": "USDC",
      "tokenOut": "WETH",
      "amountIn": "100",
      "amountOut": "0.05189"
    },
    {
      "protocol": "PancakeSwap V3",
      "poolAddress": "0x1b02dA8Cb0d097eB8D57A175b88c7D8b47997506",
      "fee": 2500,
      "tokenIn": "USDC",
      "tokenOut": "WETH",
      "amountIn": "100",
      "amountOut": "0.05162"
    }
  ]
}
```

## Error Responses

### Invalid Request Format

```json
{
  "error": "Invalid request format: binding error"
}
```

### Missing Required Parameter

```json
{
  "error": "Missing required parameter: tokena"
}
```

### Invalid Ethereum Address

```json
{
  "error": "Invalid Ethereum address format"
}
```

### Failed to Fetch Token Metadata

```json
{
  "error": "Failed to fetch token metadata: contract call error"
}
```

### Authentication Error

```json
{
  "error": "Missing API key"
}
```

```json
{
  "error": "Invalid API key"
}
```

### Internal Server Error

```json
{
  "error": "Internal server error occurred"
}
```

## Example Requests

### Get Token Information

```bash
curl --location 'http://localhost:8080/token?address=0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D'
```

### Save Token Information

```bash
curl --location 'http://localhost:8080/token' \
--header 'x-api-key: your-api-key' \
--header 'Content-Type: application/json' \
--data '{
  "address": "0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D"
}'
```

### Find Best Trading Route

```bash
curl --location 'http://localhost:8080/pairs?tokena=0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D&tokenb=0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2&amount=100'
```

### Get All Trading Routes

```bash
curl --location 'http://localhost:8080/pairs?tokena=0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D&tokenb=0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2&amount=100&all=true'
```

## Description

This project is a decentralized finance (DeFi) aggregator that allows users to view and compare DeFi projects and find the best trading route.

## Requirements

## Install Go

```bash

```

## Copy system 

## Starting

Start the Redis server:

```bash
docker compose up
```
