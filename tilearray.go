package main

/*
typedef unsigned int GLuint;
typedef unsigned int GLenum;
typedef int GLint;
typedef int GLsizei;
typedef void (*glproc)(void);

// Provided by the GLFW that raylib-go bundles.
extern glproc glfwGetProcAddress(const char *name);

static void (*pGenTextures)(GLsizei, GLuint *);
static void (*pDeleteTextures)(GLsizei, const GLuint *);
static void (*pBindTexture)(GLenum, GLuint);
static void (*pActiveTexture)(GLenum);
static void (*pTexImage3D)(GLenum, GLint, GLint, GLsizei, GLsizei, GLsizei, GLint, GLenum, GLenum, const void *);
static void (*pTexParameteri)(GLenum, GLenum, GLint);
static void (*pGenerateMipmap)(GLenum);

#define LOAD(fn, name) \
	fn = (__typeof__(fn))glfwGetProcAddress(name); \
	if (!fn) return 0;

static int loadGL(void) {
	LOAD(pGenTextures, "glGenTextures")
	LOAD(pDeleteTextures, "glDeleteTextures")
	LOAD(pBindTexture, "glBindTexture")
	LOAD(pActiveTexture, "glActiveTexture")
	LOAD(pTexImage3D, "glTexImage3D")
	LOAD(pTexParameteri, "glTexParameteri")
	LOAD(pGenerateMipmap, "glGenerateMipmap")
	return 1;
}

#define TEXTURE_2D_ARRAY 0x8C1A
#define TEXTURE0 0x84C0
#define TEXTURE_MIN_FILTER 0x2801
#define TEXTURE_MAG_FILTER 0x2800
#define TEXTURE_WRAP_S 0x2802
#define TEXTURE_WRAP_T 0x2803
#define NEAREST 0x2600
#define NEAREST_MIPMAP_LINEAR 0x2702
#define CLAMP_TO_EDGE 0x812F
#define RGBA 0x1908
#define RGBA8 0x8058
#define UNSIGNED_BYTE 0x1401

static GLuint createTileArray(int w, int h, int layers, const unsigned char *pixels) {
	if (!loadGL()) return 0;
	GLuint id = 0;
	pGenTextures(1, &id);
	pBindTexture(TEXTURE_2D_ARRAY, id);
	pTexImage3D(TEXTURE_2D_ARRAY, 0, RGBA8, w, h, layers, 0, RGBA, UNSIGNED_BYTE, pixels);
	pTexParameteri(TEXTURE_2D_ARRAY, TEXTURE_MIN_FILTER, NEAREST_MIPMAP_LINEAR);
	pTexParameteri(TEXTURE_2D_ARRAY, TEXTURE_MAG_FILTER, NEAREST);
	pTexParameteri(TEXTURE_2D_ARRAY, TEXTURE_WRAP_S, CLAMP_TO_EDGE);
	pTexParameteri(TEXTURE_2D_ARRAY, TEXTURE_WRAP_T, CLAMP_TO_EDGE);
	pGenerateMipmap(TEXTURE_2D_ARRAY);
	pBindTexture(TEXTURE_2D_ARRAY, 0);
	return id;
}

// Texture unit 0 keeps separate 2D and array bindings, so raylib's own binding is left alone.
static void bindTileArray(GLuint id) {
	pActiveTexture(TEXTURE0);
	pBindTexture(TEXTURE_2D_ARRAY, id);
}

static void deleteTileArray(GLuint id) {
	pDeleteTextures(1, &id);
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// TileArray is an OpenGL array texture with one layer per block tile.
// Layers are filtered and mipmapped on their own, so tiles never bleed into each other.
type TileArray struct{ id C.GLuint }

// NewTileArray uploads RGBA layers of size*size pixels each. The window must exist.
func NewTileArray(layers []byte, size, count int) (TileArray, error) {
	if len(layers) != size*size*4*count {
		return TileArray{}, errors.New("tile array: pixel data does not match the layer count")
	}
	id := C.createTileArray(C.int(size), C.int(size), C.int(count), (*C.uchar)(unsafe.Pointer(&layers[0])))
	if id == 0 {
		return TileArray{}, errors.New("tile array: OpenGL array textures are not available")
	}
	return TileArray{id}, nil
}

// Bind makes the array the sampler2DArray source on texture unit 0.
func (t TileArray) Bind() { C.bindTileArray(t.id) }

func (t TileArray) Unload() { C.deleteTileArray(t.id) }
