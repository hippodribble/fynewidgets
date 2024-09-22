# Some Widgets

Because large images can take time to render on a device, especially when the image has more pixels than the display device, it makes sense to display an image with as few pixels as possible.

These widgets approach this by using multiple images of varying resolution.

## PanZoom Widget

This widget contains a Gaussian pyramid of lower-resolution images. For very large images, this allows the user to zoom in and out smoothly and quickly. The pyramid handling is automatic.

The mouse wheel can be used to zoom in and out smoothly, in adjustable increments. Panning is achieved by zooming out at one position, and zooming in at another.

Additionally, the widget provides several useful hooks:
1. Loupe - a small window at full-resolution that tracks the mouse. ``DetailedImage(width,height int) `` returns a pointer to the image that can be used in an application
2. Full screen request - so that a UI can be notified to hide other containers, etc, to increase the size of the widget. Left click to activate, Attach to ``Requestfullscreen binding.Bool`` to use it.
3. Message Channel. This can be used to report information to an application's status bar, for example. Read the the ``PanZoomWidget.messages chan string`` to get updates.

Below is a screen grab - there is a thumbnail at top left, and a loupe at bottom left. The image itself is over 700 megapixels. At 100% scaling, cracks in the Sun are visible. Zooming with the mouse from very large to very small is smooth.

![Van Gogh](images/PanZoomWidgetScreenshot.png)


## Adaptive Image Widget

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