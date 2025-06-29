package interaction

type JsonView interface {
	GetView(input JsonInteractionHandlerInput) (*JsonRepresentation, error)
}
