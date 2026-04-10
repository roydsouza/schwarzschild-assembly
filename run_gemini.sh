#!/bin/bash
export NODE_OPTIONS="-r $(pwd)/control-panel/node_modules/punycode"
gemini "$@"
