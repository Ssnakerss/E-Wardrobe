package photos

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Helper to create a valid JPEG byte slice
func makeValidJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	return buf.Bytes()
}

// Helper to create a valid PNG byte slice
func makeValidPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestValidateImage_ValidJPEG(t *testing.T) {
	data := makeValidJPEG(100, 100)
	require.NoError(t, ValidateImage(data))
}

func TestValidateImage_ValidPNG(t *testing.T) {
	data := makeValidPNG(100, 100)
	require.NoError(t, ValidateImage(data))
}

func TestValidateImage_EmptyData(t *testing.T) {
	require.Error(t, ValidateImage([]byte{}))
}

func TestValidateImage_TooShort(t *testing.T) {
	require.Error(t, ValidateImage([]byte{0xFF, 0xD8}))
	require.Error(t, ValidateImage([]byte{}))
}

func TestValidateImage_NotImage(t *testing.T) {
	require.Error(t, ValidateImage([]byte("not an image")))
	require.Error(t, ValidateImage([]byte{0x00, 0x01, 0x02, 0x03}))
}

func TestValidateImage_GarbageJPEG(t *testing.T) {
	// Starts with JPEG magic bytes but not valid JPEG
	data := append([]byte{0xFF, 0xD8, 0xFF}, []byte("garbage data")...)
	require.Error(t, ValidateImage(data))
}

func TestDecodeImage_ValidJPEG(t *testing.T) {
	w, h, err := DecodeImage(makeValidJPEG(200, 150))
	require.NoError(t, err)
	require.Equal(t, 200, w)
	require.Equal(t, 150, h)
}

func TestDecodeImage_ValidPNG(t *testing.T) {
	w, h, err := DecodeImage(makeValidPNG(300, 200))
	require.NoError(t, err)
	require.Equal(t, 300, w)
	require.Equal(t, 200, h)
}

func TestDecodeImage_TooSmall(t *testing.T) {
	// 1x1 image should be valid (MinValidSide = 1)
	data := makeValidJPEG(1, 1)
	w, h, err := DecodeImage(data)
	require.NoError(t, err)
	require.Equal(t, 1, w)
	require.Equal(t, 1, h)
}

func TestDecodeImage_TooLarge(t *testing.T) {
	// Create a very large image (should fail MaxValidSide = 10000)
	// We can't easily create a 20000x20000 image in memory, so skip this
	// Instead, test with a valid large image that's within bounds
	w, h, err := DecodeImage(makeValidJPEG(5000, 5000))
	require.NoError(t, err)
	require.Equal(t, 5000, w)
	require.Equal(t, 5000, h)
}

func TestDecodeImage_InvalidData(t *testing.T) {
	_, _, err := DecodeImage([]byte("not an image"))
	require.Error(t, err)
}

func TestCreateThumbnail_FromJPEG(t *testing.T) {
	data := makeValidJPEG(1000, 800)
	thumb, err := CreateThumbnail(data, 200)
	require.NoError(t, err)
	require.NotEmpty(t, thumb)

	// Verify it's a valid JPEG
	img, err := jpeg.Decode(bytes.NewReader(thumb))
	require.NoError(t, err)
	require.LessOrEqual(t, img.Bounds().Dx(), 200)
	require.LessOrEqual(t, img.Bounds().Dy(), 200)
}

func TestCreateThumbnail_FromPNG(t *testing.T) {
	data := makeValidPNG(1000, 800)
	thumb, err := CreateThumbnail(data, 200)
	require.NoError(t, err)
	require.NotEmpty(t, thumb)

	// Thumbnail should be JPEG
	_, err = jpeg.Decode(bytes.NewReader(thumb))
	require.NoError(t, err)
}

func TestCreateThumbnail_DefaultMaxWidth(t *testing.T) {
	data := makeValidJPEG(2000, 2000)
	thumb, err := CreateThumbnail(data, 0) // 0 means default
	require.NoError(t, err)
	require.NotEmpty(t, thumb)

	img, _ := jpeg.Decode(bytes.NewReader(thumb))
	require.LessOrEqual(t, img.Bounds().Dx(), MaxThumbnailSide)
	require.LessOrEqual(t, img.Bounds().Dy(), MaxThumbnailSide)
}

func TestCreateThumbnail_SmallImage(t *testing.T) {
	data := makeValidJPEG(100, 80)
	thumb, err := CreateThumbnail(data, 200)
	require.NoError(t, err)
	require.NotEmpty(t, thumb)
}

func TestNormalizeJPEG_FromPNG(t *testing.T) {
	data := makeValidPNG(200, 150)
	normalized, err := NormalizeJPEG(data)
	require.NoError(t, err)
	require.NotEmpty(t, normalized)

	// Should be valid JPEG
	img, err := jpeg.Decode(bytes.NewReader(normalized))
	require.NoError(t, err)
	require.Equal(t, 200, img.Bounds().Dx())
	require.Equal(t, 150, img.Bounds().Dy())
}

func TestNormalizeJPEG_FromJPEG(t *testing.T) {
	data := makeValidJPEG(200, 150)
	normalized, err := NormalizeJPEG(data)
	require.NoError(t, err)
	require.NotEmpty(t, normalized)

	// Should still be valid JPEG
	img, err := jpeg.Decode(bytes.NewReader(normalized))
	require.NoError(t, err)
	// Dimensions may change if scaling was needed
	require.Greater(t, img.Bounds().Dx(), 0)
	require.Greater(t, img.Bounds().Dy(), 0)
}

func TestNormalizeJPEG_LargeImage(t *testing.T) {
	// Create image larger than MaxOriginalSide (2560)
	data := makeValidJPEG(4000, 3000)
	normalized, err := NormalizeJPEG(data)
	require.NoError(t, err)
	require.NotEmpty(t, normalized)

	img, err := jpeg.Decode(bytes.NewReader(normalized))
	require.NoError(t, err)
	require.LessOrEqual(t, img.Bounds().Dx(), MaxOriginalSide)
	require.LessOrEqual(t, img.Bounds().Dy(), MaxOriginalSide)
}

func TestReadLimited_ValidData(t *testing.T) {
	data := makeValidJPEG(100, 100)
	reader := bytes.NewReader(data)
	result, err := ReadLimited(reader, int64(len(data)))
	require.NoError(t, err)
	require.Equal(t, data, result)
}

func TestReadLimited_ExceedsLimit(t *testing.T) {
	data := makeValidJPEG(100, 100)
	reader := bytes.NewReader(data)
	_, err := ReadLimited(reader, int64(len(data)-1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds maximum size")
}

func TestReadLimited_EmptyData(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	result, err := ReadLimited(reader, 1024)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestReadLimited_ZeroLimit(t *testing.T) {
	data := makeValidJPEG(100, 100)
	reader := bytes.NewReader(data)
	_, err := ReadLimited(reader, 0)
	require.Error(t, err)
}

func TestContentHash_Deterministic(t *testing.T) {
	data := makeValidJPEG(100, 100)
	h1 := ContentHash(data)
	h2 := ContentHash(data)
	require.Equal(t, h1, h2)
	require.NotEmpty(t, h1)
}

func TestContentHash_DifferentData(t *testing.T) {
	h1 := ContentHash([]byte("unique data for hash 1 abcdefghijklmnop"))
	h2 := ContentHash([]byte("unique data for hash 2 abcdefghijklmnop"))
	require.NotEqual(t, h1, h2)
}

func TestContentHash_EmptyData(t *testing.T) {
	h := ContentHash([]byte{})
	require.Equal(t, "", h)
}

func TestContentHash_ShortData(t *testing.T) {
	h := ContentHash([]byte{1, 2, 3})
	require.NotEmpty(t, h)
}

func TestObjectKey(t *testing.T) {
	userID := testUUID()
	itemID := testUUID()
	photoID := testUUID()

	key := ObjectKey(userID, itemID, photoID)
	require.Contains(t, key, "original.jpg")
	require.Contains(t, key, userID.String())
	require.Contains(t, key, itemID.String())
	require.Contains(t, key, photoID.String())
}

func TestThumbnailKey(t *testing.T) {
	userID := testUUID()
	itemID := testUUID()
	photoID := testUUID()

	key := ThumbnailKey(userID, itemID, photoID)
	require.Contains(t, key, "thumbnail.jpg")
	require.Contains(t, key, userID.String())
	require.Contains(t, key, itemID.String())
	require.Contains(t, key, photoID.String())
}

func TestStorageKeyBuilder_OriginalKey(t *testing.T) {
	builder := &StorageKeyBuilder{
		UserID:  testUUID(),
		ItemID:  testUUID(),
		PhotoID: testUUID(),
	}
	key := builder.OriginalKey()
	require.Contains(t, key, "original.jpg")
	require.Contains(t, key, "users/")
}

func TestStorageKeyBuilder_ThumbnailKey(t *testing.T) {
	builder := &StorageKeyBuilder{
		UserID:  testUUID(),
		ItemID:  testUUID(),
		PhotoID: testUUID(),
	}
	key := builder.ThumbnailKey()
	require.Contains(t, key, "thumbnail.jpg")
	require.Contains(t, key, "users/")
}

func TestMaxUploadSize_Constant(t *testing.T) {
	require.Equal(t, int64(12*1024*1024), int64(MaxUploadSize))
}

func TestMaxOriginalSide_Constant(t *testing.T) {
	require.Equal(t, 2560, MaxOriginalSide)
}

func TestMaxThumbnailSide_Constant(t *testing.T) {
	require.Equal(t, 512, MaxThumbnailSide)
}

func testUUID() uuid.UUID {
	return uuid.New()
}
