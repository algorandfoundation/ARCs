import { ApprovalOption, ScopeType } from "./types.js"
import * as readlineSync from 'readline-sync'

export const promptUser = (bytes: Uint8Array, signer: Uint8Array, message: string): string => {
    console.log(message)
    
    const options = [ApprovalOption.CONFIRM, ApprovalOption.REJECT]
    const index = readlineSync.keyInSelect(options, 'Choose an option:', { cancel: false })
  
    if (index === -1) {
      // User canceled the prompt
      process.exit(0)
    }
  
    return options[index]
  }

export const signingMessage = (scope: ScopeType, signer: Uint8Array, bytes: Uint8Array): string => {
  const signerStr = Buffer.from(signer).toString("base64")
  return `You are about to sign bytes:${bytes}\nfor the scope: ${ScopeType[scope]}\nwith: ${signerStr}`
}