#version 330 core
in vec2 TexCoords;
in vec3 SpriteColor;
out vec4 color;

uniform sampler2D image;
void main()
{
    color = vec4(SpriteColor, 1.0) * texture(image, TexCoords);
}
