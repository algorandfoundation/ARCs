import { ApprovalOption } from "./types.js"
import * as readlineSync from 'readline-sync'

export const promptUser = (signer: string, message?: string): string => {
    console.log(`You are about to sign bytes with: ${signer}`)
    if (message) {
      console.log(`Signing message: ${message}` )
    }
    
    const options = [ApprovalOption.CONFIRM, ApprovalOption.REJECT]
    const index = readlineSync.keyInSelect(options, 'Choose an option:', { cancel: false })
  
    if (index === -1) {
      // User canceled the prompt
      process.exit(0)
    }
  
    return options[index]
  }