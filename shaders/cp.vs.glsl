#version 330 core
in vec2 vertex;
in vec2 aa_coord;
in vec4 fill_color;
in vec4 outline_color;

out vec2 v_aa_coord;
out vec4 v_fill_color;
out vec4 v_outline_color;

uniform mat4 projection;

void main(void){
    gl_Position = projection*vec4(vertex.xy, 0.0, 1.0);

    v_fill_color = fill_color;
    v_outline_color = outline_color;
    v_aa_coord = aa_coord;
}
