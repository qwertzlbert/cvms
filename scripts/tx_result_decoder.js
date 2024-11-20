#!/usr/bin/env node
// modules
const fs = require('fs');
const path = require('path');

// functions
const decodeBase64IfPossible = function (text) {
  if (text === null) {
    return null;
  }

  try {
    const decodedText = Buffer.from(text, 'base64').toString('utf8');
    // Check if the decoded string, when re-encoded to base64, matches the original text
    if (Buffer.from(decodedText).toString('base64') === text) {
      return decodedText;
    }
  } catch (error) {
    // If an error occurs, return the input text as is
  }

  return text;
};
const decodeAttributes = function (attributes) {
  return attributes.map((attr) => {
    return {
      ...attr,
      key: decodeBase64IfPossible(attr.key),
      value: decodeBase64IfPossible(attr.value),
    };
  });
};

// Get the file name from the command-line arguments
const inputFilePath = process.argv[2]; // First argument after `node script.js`
if (!inputFilePath) {
  console.error('Please provide an input file name as an argument.');
  process.exit(1); // Exit the process with a failure code
}

const absolutePath = path.resolve(inputFilePath);

// Extract the directory name and base name
const dirName = path.dirname(absolutePath); // Get the directory name
const fileName = path.basename(absolutePath); // Get the file name including extension
// Construct the output file name and path

const outputFileName = `decoded_${fileName}`; // Add a prefix to the base name
const outputFilePath = path.join(dirName, outputFileName); // Combine directory and output file name

// Try decoding
try {
  const jsonData = JSON.parse(fs.readFileSync(inputFilePath, 'utf-8'));
  jsonData.result.txs_results.forEach((tx) => {
    tx.events.forEach((event) => {
      event.attributes = decodeAttributes(event.attributes);
    });
  });
  jsonData.result.begin_block_events.forEach((event) => {
    event.attributes = decodeAttributes(event.attributes);
  });

  // Write processed JSON to the output file
  fs.writeFileSync(outputFilePath, JSON.stringify(jsonData, null, 2), 'utf-8');
  console.log(`Decoded JSON data saved to ${outputFilePath}`);
} catch (err) {
  console.error('Error reading or processing the file:', err.message);
  process.exit(1); // Exit the process with a failure code
}
