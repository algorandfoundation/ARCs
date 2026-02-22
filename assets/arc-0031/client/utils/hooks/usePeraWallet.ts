import type { PeraWalletArbitraryData } from '@perawallet/connect/dist/util/model/peraWalletModels'

export const usePeraWallet = () => {
  const { $peraWalletClient } = useNuxtApp()

  const address = useState('address', () => '')

  const disconnectWallet = async () => {
    if (!$peraWalletClient.isConnected) {
      return
    }

    await $peraWalletClient.disconnect()
    address.value = ''
  }

  const connectWallet = async (
  ) => {
    const accounts = await $peraWalletClient.connect()
    $peraWalletClient.connector?.on('disconnect', disconnectWallet)
    address.value = accounts[0]

    return address
  }

  const reconnectWallet = async () => {
    if ($peraWalletClient.isConnected) {
      return
    }

    const accounts = await $peraWalletClient.reconnectSession()
    if (accounts.length) {
      $peraWalletClient.connector?.on('disconnect', disconnectWallet)
      address.value = accounts[0]
    }
  }

  const signData = (data: PeraWalletArbitraryData[], address: string) =>
    $peraWalletClient.signData(data, address)

  return {
    peraWalletClient: $peraWalletClient,
    address,
    connectWallet,
    disconnectWallet,
    reconnectWallet,
    signData
  }
}
