import ctypes
import pygame
import math
import sys

# custom flag creation
# print([ord(c) for c in "FLAG{YOUR_CUSTOM_FLAG}"])

# Constants
WIDTH, HEIGHT = 800, 800
FPS = 60
# Finish line as vertical line
FINISH_LINE_X = 350
FINISH_LINE_Y1 = 650
FINISH_LINE_Y2 = 750
# Checkpoint line coordinates
CHECK_X = 300
CHECK_Y1 = 50
CHECK_Y2 = 200

# Speed dynamics
BASE_SPEED = 2.5      # px per frame at rest
MAX_SPEED = 5.0       # px per frame max
ACCELERATION = 0.005  # speed units per ms when straight
DECELERATION = 0.01   # speed units per ms when turning
# Conversion to fake km/h
KMH_SCALE = 30        # 1 speed-unit ≈ 20 km/h

# Best personal record (seconds)
BEST_PR = 10.0

class Car:
    def __init__(self, image_path, x, y, angle=0):
        self.original_image = pygame.image.load(image_path).convert_alpha()
        self.start_x = x
        self.start_y = y
        self.start_angle = angle
        self.reset()

    def reset(self):
        self.x = self.start_x
        self.y = self.start_y
        self.angle = self.start_angle
        self.speed = BASE_SPEED

    def update(self, track_mask, dt, turning):
        # Adjust speed: accelerate when straight, decelerate when turning
        if not turning:
            self.speed = min(MAX_SPEED, self.speed + ACCELERATION * dt)
        else:
            self.speed = max(BASE_SPEED, self.speed - DECELERATION * dt)

        # Move car
        rad = math.radians(self.angle)
        dx = math.cos(rad) * self.speed
        dy = math.sin(rad) * self.speed
        new_x = self.x + dx
        new_y = self.y + dy

        # Only move if on road (mask white)
        ix, iy = int(new_x), int(new_y)
        if 0 <= ix < WIDTH and 0 <= iy < HEIGHT and track_mask.get_at((ix, iy))[0] > 128:
            self.x, self.y = new_x, new_y

    def draw(self, surface):
        rotated = pygame.transform.rotate(self.original_image, -self.angle)
        rect = rotated.get_rect(center=(self.x, self.y))
        surface.blit(rotated, rect)

# Obfuscated flag builder
def _():
    __= [78, 73, 67, 69, 95, 82, 65, 67, 69, 95, 82, 79, 79, 75, 73, 69]
    return ''.join(map(chr, __))


def run_game():
    pygame.init()
    screen = pygame.display.set_mode((WIDTH, HEIGHT))
    pygame.display.set_caption("TURBO TIMER")
    clock = pygame.time.Clock()
    font = pygame.font.SysFont(None, 36)

    # Load assets
    track = pygame.image.load('assets/track.png').convert()
    track_mask = pygame.image.load('assets/track_mask.png').convert()
    player = Car('assets/red_car.png', FINISH_LINE_X + 50, 700, angle=180)

    global BEST_PR
    while True:
        # Use ctypes float to store game time in memory (easy to find with CE)
        game_time = ctypes.c_float(0.0)

        finished = False
        finish_time = 0
        checkpoint_passed = False
        player.reset()
        old_pr = BEST_PR

        # Main trial loop
        while not finished:
            dt = clock.tick(FPS)
            game_time.value += dt / 1000.0  # dt is in milliseconds
            turning = False

            for event in pygame.event.get():
                if event.type == pygame.QUIT:
                    pygame.quit(); sys.exit()

            keys = pygame.key.get_pressed()
            if keys[pygame.K_LEFT]: player.angle -= 2.5; turning = True
            if keys[pygame.K_RIGHT]: player.angle += 2.5; turning = True

            player.update(track_mask, dt, turning)
            if not checkpoint_passed and player.x <= CHECK_X and CHECK_Y1 <= player.y <= CHECK_Y2:
                checkpoint_passed = True
            if checkpoint_passed and player.x <= FINISH_LINE_X and FINISH_LINE_Y1 <= player.y <= FINISH_LINE_Y2:
                finish_time = game_time.value
                finished = True

            screen.blit(track, (0,0))
            player.draw(screen)
            # Debug lines (uncomment to see)
            pygame.draw.line(screen, (255,255,0), (CHECK_X, CHECK_Y1), (CHECK_X, CHECK_Y2), 5)
            pygame.draw.line(screen, (255,0,0), (FINISH_LINE_X, FINISH_LINE_Y1), (FINISH_LINE_X, FINISH_LINE_Y2), 5)

            elapsed = finish_time if finished else game_time.value
            screen.blit(font.render(f"Time: {elapsed:.2f}s", True, (255,255,255)), (10, 10))
            screen.blit(font.render(f"Speed: {player.speed * KMH_SCALE:.0f} km/h", True, (255,255,255)), (10, 50))
            pygame.display.flip()

        # End overlay
        overlay = pygame.Surface((WIDTH, HEIGHT), pygame.SRCALPHA)
        overlay.fill((50,50,50,180))
        screen.blit(overlay, (0,0))
        BEST_PR = min(BEST_PR, finish_time)

        screen.blit(font.render(f"Best PR: {BEST_PR:.2f}s", True, (255,255,255)), (WIDTH//2-100, HEIGHT//2-60))
        screen.blit(font.render(f"Your Time: {finish_time:.2f}s", True, (255,255,255)), (WIDTH//2-100, HEIGHT//2))

        if finish_time < old_pr:
            screen.blit(font.render(_(), True, (0,255,0)), (WIDTH//2-100, HEIGHT//2+40))

        screen.blit(font.render("Press R or click to replay", True, (255,255,255)), (WIDTH//2-140, HEIGHT//2+100))
        pygame.display.flip()

        # Wait for replay
        while True:
            for event in pygame.event.get():
                if event.type == pygame.QUIT: pygame.quit(); sys.exit()
                if event.type == pygame.KEYDOWN and event.key == pygame.K_r: break
                if event.type == pygame.MOUSEBUTTONDOWN: break
            else:
                clock.tick(15)
                continue
            break


def main():
    run_game()

if __name__ == '__main__':
    main()
