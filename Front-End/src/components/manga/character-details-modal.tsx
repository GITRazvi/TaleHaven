import React from "react";
import { Alert, Col, Modal, Row, Spinner } from "react-bootstrap";
import { AniListCharacterDetails } from "../../app/clients/backend/types";

type CharacterDetailsModalProps = {
  characterID: number | null;
  character: AniListCharacterDetails | null;
  loading: boolean;
  error: string;
  onHide: () => void;
  onSelectManga: (anilistID: number) => void;
};

const characterNameOf = (character: Pick<AniListCharacterDetails, "name">) =>
  character.name.full ||
  character.name.userPreferred ||
  character.name.native ||
  "Unnamed character";

const characterDateLabel = (date: AniListCharacterDetails["dateOfBirth"]) =>
  [date.year, date.month, date.day].filter(Boolean).join("-");

const mediaTitleOf = (
  media: AniListCharacterDetails["media"]["edges"][number]["node"],
) =>
  media.title.english ||
  media.title.userPreferred ||
  media.title.romaji ||
  media.title.native ||
  "Untitled work";

const characterRoleLabel = (role: string) =>
  role
    ? role.charAt(0) + role.slice(1).toLowerCase().replace(/_/g, " ")
    : "Character";

const CharacterDetailsModal: React.FC<CharacterDetailsModalProps> = ({
  characterID,
  character,
  loading,
  error,
  onHide,
  onSelectManga,
}) => (
  <Modal show={characterID !== null} onHide={onHide} size="lg" centered>
    <Modal.Header closeButton>
      <Modal.Title>
        {character ? characterNameOf(character) : "Character details"}
      </Modal.Title>
    </Modal.Header>
    <Modal.Body>
      {loading && (
        <div className="text-center py-4">
          <Spinner animation="border" size="sm" className="me-2" />
          Loading character details...
        </div>
      )}
      {error && <Alert variant="danger" className="mb-0">{error}</Alert>}
      {character && (
        <>
          <div className="character-details-summary">
            {(character.image.large || character.image.medium) && (
              <img
                src={character.image.large || character.image.medium}
                alt={characterNameOf(character)}
                referrerPolicy="no-referrer"
              />
            )}
            <div>
              <dl className="character-details-facts mb-0">
                {character.gender && <><dt>Gender</dt><dd>{character.gender}</dd></>}
                {characterDateLabel(character.dateOfBirth) && <><dt>Born</dt><dd>{characterDateLabel(character.dateOfBirth)}</dd></>}
                {character.age && <><dt>Age</dt><dd>{character.age}</dd></>}
                {character.bloodType && <><dt>Blood type</dt><dd>{character.bloodType}</dd></>}
                {character.favourites > 0 && <><dt>AniList favorites</dt><dd>{character.favourites.toLocaleString()}</dd></>}
              </dl>
              {character.siteUrl && (
                <a href={character.siteUrl} target="_blank" rel="noreferrer">
                  View on AniList
                </a>
              )}
            </div>
          </div>
          {character.description && (
            <section className="mt-4">
              <h3 className="h5">About</h3>
              <p className="mb-0" style={{ whiteSpace: "pre-line" }}>
                {character.description}
              </p>
            </section>
          )}
          <section className="mt-4">
            <h3 className="h5">Appears in</h3>
            {character.media.edges.length === 0 ? (
              <p className="text-muted mb-0">
                AniList does not list associated works for this character.
              </p>
            ) : (
              <Row className="g-3">
                {character.media.edges.map((edge) => {
                  const workTitle = mediaTitleOf(edge.node);
                  const workCover = edge.node.coverImage.large || edge.node.coverImage.medium;
                  const canOpenManga = edge.node.type === "MANGA";
                  return (
                    <Col key={edge.node.id} sm={6}>
                      <button
                        type="button"
                        className="character-media-item"
                        onClick={() => canOpenManga && onSelectManga(edge.node.id)}
                        disabled={!canOpenManga}
                      >
                        {workCover ? (
                          <img src={workCover} alt={workTitle} referrerPolicy="no-referrer" />
                        ) : (
                          <div className="character-media-placeholder" aria-hidden="true">?</div>
                        )}
                        <span>
                          <strong>{workTitle}</strong>
                          <small>
                            {canOpenManga ? "Manga" : edge.node.type}
                            {edge.characterRole ? ` · ${characterRoleLabel(edge.characterRole)}` : ""}
                          </small>
                        </span>
                      </button>
                    </Col>
                  );
                })}
              </Row>
            )}
          </section>
        </>
      )}
    </Modal.Body>
  </Modal>
);

export default CharacterDetailsModal;
