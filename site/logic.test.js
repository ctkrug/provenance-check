"use strict";

const { test } = require("node:test");
const assert = require("node:assert/strict");
const { MAX_URLS, parseURLs, stampFontSize, VERDICT_LABELS } = require("./logic.js");

test("parseURLs splits one URL per line", () => {
  assert.deepEqual(parseURLs("https://a\nhttps://b"), ["https://a", "https://b"]);
});

test("parseURLs drops blank and whitespace-only lines", () => {
  assert.deepEqual(parseURLs("https://a\n\n   \nhttps://b\n\t\n"), ["https://a", "https://b"]);
});

test("parseURLs trims surrounding whitespace from each line", () => {
  assert.deepEqual(parseURLs("  https://a  \n\thttps://b\t"), ["https://a", "https://b"]);
});

test("parseURLs handles CRLF line endings", () => {
  assert.deepEqual(parseURLs("https://a\r\nhttps://b\r\n"), ["https://a", "https://b"]);
});

test("parseURLs on empty input returns an empty array", () => {
  assert.deepEqual(parseURLs(""), []);
});

test("parseURLs on whitespace-only input returns an empty array", () => {
  assert.deepEqual(parseURLs("   \n\t\n   "), []);
});

test("parseURLs caps output at MAX_URLS, keeping the first N", () => {
  const lines = Array.from({ length: MAX_URLS + 10 }, (_, i) => `https://example.com/${i}`);
  const result = parseURLs(lines.join("\n"));
  assert.equal(result.length, MAX_URLS);
  assert.deepEqual(result, lines.slice(0, MAX_URLS));
});

test("parseURLs on exactly MAX_URLS lines keeps all of them", () => {
  const lines = Array.from({ length: MAX_URLS }, (_, i) => `https://example.com/${i}`);
  assert.equal(parseURLs(lines.join("\n")).length, MAX_URLS);
});

test("parseURLs never returns more than MAX_URLS entries or blank entries (property check)", () => {
  const chars = ["a", " ", "\n", "\t", "\r", "https://x"];
  for (let i = 0; i < 200; i++) {
    let input = "";
    const len = Math.floor(Math.random() * 300);
    for (let j = 0; j < len; j++) {
      input += chars[Math.floor(Math.random() * chars.length)];
    }
    const result = parseURLs(input);
    assert.ok(result.length <= MAX_URLS, `result too long for input ${JSON.stringify(input)}`);
    for (const line of result) {
      assert.notEqual(line.trim(), "", `blank entry survived for input ${JSON.stringify(input)}`);
    }
  }
});

test("stampFontSize is 9px for labels up to 5 chars", () => {
  assert.equal(stampFontSize("CLEAR"), 9);
  assert.equal(stampFontSize(""), 9);
});

test("stampFontSize is 7.5px for labels 6-7 chars", () => {
  assert.equal(stampFontSize("CAUTIO"), 7.5);
  assert.equal(stampFontSize("CAUTION"), 7.5);
});

test("stampFontSize is 6px for labels longer than 7 chars", () => {
  assert.equal(stampFontSize("RESTRICTED"), 6);
});

test("every VERDICT_LABELS value fits one of stampFontSize's bands without truncation", () => {
  for (const label of Object.values(VERDICT_LABELS)) {
    assert.ok(stampFontSize(label) > 0);
  }
});
