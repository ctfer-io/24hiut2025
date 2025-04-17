#!/bin/bash

# Create the images directory if it doesn't exist
mkdir -p images

# URL of the API endpoint
api_url="https://api.imgflip.com/get_memes"

# Fetch the JSON data from the API
json_data=$(curl -s "$api_url")

# Debug: Print the JSON data to verify it's fetched correctly
echo "Fetched JSON data:"
echo "$json_data"
echo

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "jq is not installed. Please install jq to parse JSON data."
    exit 1
fi

# Extract image URLs using jq
image_urls=$(echo "$json_data" | jq -r '.data.memes[].url')

# Debug: Print the extracted image URLs
echo "Extracted image URLs:"
echo "$image_urls"
echo

# Check if image_urls is empty
if [ -z "$image_urls" ]; then
    echo "No image URLs were extracted. Please check the JSON data and jq command."
    exit 1
fi

# Download each image
for url in $image_urls; do
    wget -P images/ "$url"
done

echo "Download complete."
