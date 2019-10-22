package fam

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type Banana struct {
	Texture *Texture2D

	*cp.Body
	*cp.Shape
	*cp.Circle
}

func NewBanana(pos cp.Vector, radius float64, sprite *Texture2D, space *cp.Space) *Banana {
	p := &Banana{
		Texture: sprite,
	}
	p.Body = cp.NewBody(1, cp.MomentForCircle(1, radius, radius, cp.Vector{0, 0}))
	p.Shape = cp.NewCircle(p.Body, radius, cp.Vector{0, 0})
	p.Shape.SetElasticity(0)
	p.Shape.SetFriction(1)

	// for consummation
	p.Shape.SetCollisionType(collisionBanana)
	p.Shape.SetFilter(PlayerFilter)

	p.Shape.UserData = p
	p.Circle = p.Shape.Class.(*cp.Circle)
	p.Body.SetPosition(pos)
	space.AddBody(p.Body)
	space.AddShape(p.Shape)
	return p
}

func (p *Banana) Update(g *Game, dt float64) {

}

func (p *Banana) Draw(renderer *SpriteRenderer, alpha float64) {
	pos := p.Position()
	//pos = pos.Mult(alpha).Add(p.LastPosition.Mult(1.0 - alpha))
	bb := p.Shape.BB()
	size := mgl32.Vec2{
		float32(bb.R - bb.L),
		float32(bb.T - bb.B),
	}
	renderer.DrawSprite(p.Texture, V(pos), size, p.Angle(), mgl32.Vec3{1, 1, 1})
}

func BananaPreSolve(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
	game := data.(*Game)

	a, b := arb.Shapes()
	banana := a.UserData.(*Banana)

	switch b.UserData.(type) {
	case *Player:
		player := b.UserData.(*Player)

		player.SetRadius(player.Radius() * 1.1)

		space.AddPostStepCallback(func(s *cp.Space, a interface{}, b interface{}) {
			if banana.Shape == nil {
				return
			}
			banana.Shape.UserData = nil
			s.RemoveShape(banana.Shape)
			s.RemoveBody(banana.Body)
			banana.Shape = nil
			banana.Body = nil
			for i := 0; i < len(game.Bananas); i++ {
				if game.Bananas[i] == banana {
					game.Bananas = append(game.Bananas[:i], game.Bananas[i+1:]...)
					return
				}
			}
		}, nil, nil)

		arb.Ignore()
		return false
	}

	return true
}
