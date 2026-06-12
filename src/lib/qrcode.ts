import QRCode from 'qrcode'
import jsQR from 'jsqr'

const QR_PREFIX = 'PCA:'
const CHECKSUM_SEED = 0xCAFE

function crc16(data: string): string {
  let crc = CHECKSUM_SEED
  for (let i = 0; i < data.length; i++) {
    crc ^= data.charCodeAt(i) << 8
    for (let j = 0; j < 8; j++) {
      crc = (crc & 0x8000) !== 0 ? (crc << 1) ^ 0x1021 : crc << 1
      crc &= 0xFFFF
    }
  }
  return crc.toString(16).toUpperCase().padStart(4, '0')
}

export function generateQRContent(cartridgeId: number): string {
  const idStr = String(cartridgeId)
  const checksum = crc16(idStr)
  return `${QR_PREFIX}${idStr}:${checksum}`
}

export function parseQRContent(content: string): {
  valid: boolean
  cartridgeId: number | null
} {
  try {
    if (!content.startsWith(QR_PREFIX)) {
      return { valid: false, cartridgeId: null }
    }
    const payload = content.slice(QR_PREFIX.length)
    const parts = payload.split(':')
    if (parts.length !== 2) {
      return { valid: false, cartridgeId: null }
    }
    const [idStr, checksum] = parts
    const expectedChecksum = crc16(idStr)
    if (checksum !== expectedChecksum) {
      return { valid: false, cartridgeId: null }
    }
    const id = parseInt(idStr, 10)
    if (isNaN(id) || id <= 0) {
      return { valid: false, cartridgeId: null }
    }
    return { valid: true, cartridgeId: id }
  } catch {
    return { valid: false, cartridgeId: null }
  }
}

export async function generateQRDataURL(
  cartridgeId: number,
  size: number = 200
): Promise<string> {
  const content = generateQRContent(cartridgeId)
  return QRCode.toDataURL(content, {
    width: size,
    margin: 2,
    color: {
      dark: '#1A1A2E',
      light: '#FFFFFF'
    }
  })
}

export async function generateQRCanvas(
  cartridgeId: number,
  canvas: HTMLCanvasElement,
  size: number = 200
): Promise<void> {
  const content = generateQRContent(cartridgeId)
  await QRCode.toCanvas(canvas, content, {
    width: size,
    margin: 2,
    color: {
      dark: '#1A1A2E',
      light: '#FFFFFF'
    }
  })
}

export function decodeQRFromImage(
  imageData: ImageData
): { valid: boolean; cartridgeId: number | null } {
  const code = jsQR(imageData.data, imageData.width, imageData.height)
  if (!code) {
    return { valid: false, cartridgeId: null }
  }
  return parseQRContent(code.data)
}

export function decodeQRFromCanvas(
  canvas: HTMLCanvasElement
): { valid: boolean; cartridgeId: number | null } {
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    return { valid: false, cartridgeId: null }
  }
  const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
  return decodeQRFromImage(imageData)
}
