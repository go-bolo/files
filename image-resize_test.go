package files

// func TestResizeImage(t *testing.T) {
// 	assert := assert.New(t)

// 	mockFilePath := "../_stubs/file/tux.png"

// 	styles := map[string]ImageStyleCfg{
// 		"thumbnail": {
// 			Width:  75,
// 			Height: 75,
// 		},
// 	}

// 	t.Run("Should resize one image", func(t *testing.T) {
// 		fileId := uuid.New().String() + ".png"
// 		tmpFilePath := path.Join(os.TempDir(), fileId)
// 		resizedPath := path.Join(os.TempDir(), "resized_"+fileId)

// 		// cleanup:
// 		defer os.Remove(tmpFilePath)
// 		defer os.Remove(resizedPath)

// 		err := files_helpers.CopyFile(mockFilePath, tmpFilePath)
// 		assert.NoError(err)

// 		style := styles["thumbnail"]

// 		err = ResizeImage(style.Width, style.Height, tmpFilePath, resizedPath)
// 		assert.NoError(err)

// 		buffer, err := bimg.Read(resizedPath)
// 		assert.NoError(err)

// 		sourceImage := bimg.NewImage(buffer)
// 		finalSize, err := sourceImage.Size()
// 		assert.NoError(err)

// 		assert.Equal(style.Height, finalSize.Height)
// 		assert.Equal(style.Width, finalSize.Width)

// 		assert.True(true)
// 	})
// }

// func TestResizeImageIfIsBiggerThan(t *testing.T) {
// 	assert := assert.New(t)

// 	mockFilePath := "../_stubs/file/tux.png"

// 	t.Run("Should resize one image", func(t *testing.T) {
// 		fileId := uuid.New().String() + ".png"
// 		tmpFilePath := path.Join(os.TempDir(), fileId)
// 		resizedPath := path.Join(os.TempDir(), "resized_"+fileId)

// 		maxWidth := uint(10)
// 		maxHeight := uint(40)

// 		// cleanup:
// 		defer os.Remove(tmpFilePath)
// 		defer os.Remove(resizedPath)

// 		err := files_helpers.CopyFile(mockFilePath, tmpFilePath)
// 		assert.NoError(err)

// 		err = ResizeImageIfIsBiggerThan(maxWidth, maxHeight, tmpFilePath, resizedPath)
// 		assert.NoError(err)

// 		buffer, err := bimg.Read(resizedPath)
// 		assert.NoError(err)

// 		sourceImage := bimg.NewImage(buffer)
// 		finalSize, err := sourceImage.Size()
// 		assert.NoError(err)

// 		assert.Equal(maxHeight, uint(finalSize.Height))
// 		assert.Equal(maxWidth, uint(finalSize.Width))

// 		assert.True(true)
// 	})
// }
