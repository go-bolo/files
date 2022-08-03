package files

func ResizeImage(width, height int, localFilePath, localDest string) error {
	// buffer, err := bimg.Read(localFilePath)
	// if err != nil {
	// 	return err
	// }

	// sourceImage := bimg.NewImage(buffer)

	// sourceSize, err := sourceImage.Size()
	// if err != nil {
	// 	return err
	// }

	// imageWidth := sourceSize.Width
	// imageHeight := sourceSize.Height

	// if imageWidth < width || imageHeight < height {
	// 	sourceEnlargedImage, err := sourceImage.Enlarge(int(width), int(height))
	// 	if err != nil {
	// 		return err
	// 	}

	// 	sourceImage = bimg.NewImage(sourceEnlargedImage)
	// }

	// newImage, err := sourceImage.ResizeAndCrop(int(width), int(height))
	// if err != nil {
	// 	return err
	// }

	// bimg.Write(localDest, newImage)

	return nil
}

func ResizeImageIfIsBiggerThan(maxWidth, maxHeight uint, localFilePath, localDest string) error {

	// buffer, err := bimg.Read(localFilePath)
	// if err != nil {
	// 	return err
	// }

	// sourceImage := bimg.NewImage(buffer)

	// sourceSize, err := sourceImage.Size()
	// if err != nil {
	// 	return err
	// }

	// imageWidth := uint(sourceSize.Width)
	// imageHeight := uint(sourceSize.Height)

	// if imageWidth <= maxWidth && imageHeight <= maxHeight {
	// 	return nil
	// }

	// newImage, err := bimg.NewImage(buffer).ResizeAndCrop(int(maxWidth), int(maxHeight))
	// if err != nil {
	// 	return err
	// }

	// bimg.Write(localDest, newImage)

	return nil
}
