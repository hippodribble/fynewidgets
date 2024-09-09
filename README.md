# Some Widgets

## Adaptive Image Widget

Because large images can take time to render on a device, especially when the image has more pixels than the display device, it makes sense to display an image with as few pixels as possible.

The AdaptiveImageWidget addresses this by replacing the image.Image in a standard canvas.Image with a Gaussian pyramid of images. This pyramid contains several layers, each at half the resolution of the previous layer, to some defined minimum size.

#### Dynamic Resizing
When it is resized, the widget checks if it is larger or smaller than the embedded image. This causes the appropriate resolution image to be selected from the pyramid and displayed.

This results in a more responsive display under resizing.

#### Memory Requirements 
memory usage for the pyramid is around 1.5 times that of the underlying image.Image, including the full-resolution image, assuming size * ( 1 + 1/4 + 1/16 + ... )

## Thumbnail Widget

Can be used to generate a thumbnail in a specified rectangle (which it will compeltely fill)

## Image Detail Widget

Large images are displayed smaller than 1 screen pixel per image pixel. This widget allows users to display a sub-image at full scale to check image detail, etc.

# Minimal Fyne Implementation

The cmd/ folder includes a fyne App that shows the widgets.

```
go run .\cmd\main.go  //from the fynewidgets directory
```

### Full Image Button 
- just loads an image normally. If a large image is loaded, then manually resizong the screen can be cumbersome

### AutoImage Button 
- loads the image but decimates into a pyramid. Resizing causes a more appropriately scaled imaged to be selected for display in the widget.
- A loupe image (small zoom window) is opened on the left. It tracks the mouse position.

### Folder
- opens a folder and generates thumbnails for all images.
- click on a thumbnail to display the full image. Loupe image is also displayed.
- click on the full image to toggle full window size display.