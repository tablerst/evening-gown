import { adminPostForm } from '@/admin/api'
import { appEnv } from '@/config/env'

export type UploadKind = 'cover' | 'hover'

export type UploadResult = {
    url: string
    objectKey: string
    contentType: string
    size: number
}

type CompressOptions = {
    maxBytes?: number
    maxEdge?: number
    minQuality?: number
    maxQuality?: number
}

const loadBitmap = async (file: File): Promise<ImageBitmap> => {
    // Prefer honoring EXIF orientation when supported.
    try {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return await (createImageBitmap as any)(file, { imageOrientation: 'from-image' })
    } catch {
        return await createImageBitmap(file)
    }
}

const canvasToWebp = (canvas: HTMLCanvasElement, quality: number): Promise<Blob> => {
    return new Promise((resolve, reject) => {
        canvas.toBlob(
            (blob) => {
                if (!blob) {
                    reject(new Error('无法生成 WebP（toBlob 返回空）'))
                    return
                }
                resolve(blob)
            },
            'image/webp',
            quality,
        )
    })
}

const clampInt = (n: number, min: number, max: number) => Math.max(min, Math.min(max, Math.floor(n)))

const stripExt = (name: string) => {
    const idx = name.lastIndexOf('.')
    if (idx <= 0) return name
    return name.slice(0, idx)
}

export const compressImageToWebpUnderLimit = async (file: File, options?: CompressOptions): Promise<File> => {
    const maxBytes = options?.maxBytes ?? appEnv.maxImageUploadBytes ?? 1048576
    const minQuality = options?.minQuality ?? 0.45
    const maxQuality = options?.maxQuality ?? 0.92

    if (file.type === 'image/webp' && file.size > 0 && file.size <= maxBytes) {
        return file
    }

    const bitmap = await loadBitmap(file)
    try {
        const srcW = bitmap.width
        const srcH = bitmap.height
        if (!srcW || !srcH) throw new Error('无法读取图片尺寸')

        let maxEdge = options?.maxEdge ?? 2560
        maxEdge = clampInt(maxEdge, 320, 8192)

        const canvas = document.createElement('canvas')
        const ctx = canvas.getContext('2d')
        if (!ctx) throw new Error('无法创建 canvas 上下文')

        // Try a few rounds: binary-search quality; if still too big, reduce dimensions.
        for (let round = 0; round < 6; round++) {
            const scale = Math.min(1, maxEdge / Math.max(srcW, srcH))
            const w = Math.max(1, Math.round(srcW * scale))
            const h = Math.max(1, Math.round(srcH * scale))

            canvas.width = w
            canvas.height = h
            ctx.clearRect(0, 0, w, h)
            ctx.drawImage(bitmap, 0, 0, w, h)

            let lo = minQuality
            let hi = maxQuality
            let best: Blob | null = null
            let smallest: Blob | null = null

            for (let i = 0; i < 8; i++) {
                const q = (lo + hi) / 2
                const blob = await canvasToWebp(canvas, q)
                if (!smallest || blob.size < smallest.size) smallest = blob

                if (blob.size <= maxBytes) {
                    best = blob
                    lo = q
                } else {
                    hi = q
                }
            }

            const finalBlob = best ?? smallest
            if (finalBlob && finalBlob.size <= maxBytes) {
                const outName = `${stripExt(file.name || 'image')}.webp`
                return new File([finalBlob], outName, { type: 'image/webp' })
            }

            // Still too large: shrink dimensions and retry.
            maxEdge = Math.max(320, Math.round(maxEdge * 0.85))
        }

        throw new Error(`无法将图片压缩到 ${(maxBytes / 1024 / 1024).toFixed(2)}MB 以内，请换一张或裁剪后再试`) 
    } finally {
        bitmap.close?.()
    }
}

export const uploadAdminImage = async (kind: UploadKind, styleNo: string, file: File): Promise<UploadResult> => {
    const fd = new FormData()
    fd.append('kind', kind)
    fd.append('styleNo', String(styleNo))
    fd.append('file', file, file.name)

    return adminPostForm<UploadResult>('/api/v1/admin/uploads/images', fd)
}
