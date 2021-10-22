import AnsiParser from './ansi-parser'

const ansiParser = new AnsiParser()

const result = ansiParser.parse("Thank you for installing \u001b[35mEJS\u001b[0m: built with the \u001b[32mJake\u001b[0m JavaScript build tool (\u001b[32mhttps://jakejs.com/\u001b[0m)")
console.log(result)
